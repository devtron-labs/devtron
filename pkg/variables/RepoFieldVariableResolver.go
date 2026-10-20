/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package variables

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"go.uber.org/zap"
)

// fullVariableReferenceRegex matches a field value that is exactly one @{{VAR_NAME}} token and nothing else.
var fullVariableReferenceRegex = regexp.MustCompile(`^@\{\{([A-Za-z0-9_]+)\}\}$`)

// IsFullVariableReference reports whether fieldValue is exactly one @{{VAR_NAME}} token, so
// callers (e.g. request validation) can skip literal-value format checks without a DB lookup.
func IsFullVariableReference(fieldValue string) bool {
	return fullVariableReferenceRegex.MatchString(fieldValue)
}

// ExtractVariableName returns the variable name if fieldValue is a full @{{VAR_NAME}}
// reference, without doing any DB lookup - used to track which entities (e.g. Git Materials)
// reference which variables, independent of whether the reference currently resolves.
func ExtractVariableName(fieldValue string) (string, bool) {
	matches := fullVariableReferenceRegex.FindStringSubmatch(fieldValue)
	if matches == nil {
		return "", false
	}
	return matches[1], true
}

// RepoFieldVariableResolver resolves a single, full @{{VAR_NAME}} reference used in the Git
// Repository URL and Container Repository fields. Only user-defined variables are supported -
// scoped variables in this version only support Global scope, which fits these fields anyway
// since neither has any app/env/cluster context at its resolution point.
type RepoFieldVariableResolver interface {
	// ResolveRepoFieldValue returns fieldValue unchanged if it isn't a variable reference,
	// or its resolved value if it is a valid one. variableName is empty when nothing was resolved.
	ResolveRepoFieldValue(scope resourceQualifiers.Scope, fieldValue string) (resolvedValue string, variableName string, err error)
}

type RepoFieldVariableResolverImpl struct {
	logger                *zap.SugaredLogger
	scopedVariableService ScopedVariableService
}

func NewRepoFieldVariableResolverImpl(logger *zap.SugaredLogger, scopedVariableService ScopedVariableService) *RepoFieldVariableResolverImpl {
	return &RepoFieldVariableResolverImpl{
		logger:                logger,
		scopedVariableService: scopedVariableService,
	}
}

func (impl *RepoFieldVariableResolverImpl) ResolveRepoFieldValue(scope resourceQualifiers.Scope, fieldValue string) (string, string, error) {
	// plain literal value, nothing to resolve
	if !strings.Contains(fieldValue, "@{{") {
		return fieldValue, "", nil
	}

	matches := fullVariableReferenceRegex.FindStringSubmatch(fieldValue)
	if matches == nil {
		// contains "@{{" but isn't a single, full @{{VAR_NAME}} token - not supported here
		return "", "", fmt.Errorf("only a single full variable reference like @{{VAR_NAME}} is supported, not partial text or multiple variables")
	}
	variableName := matches[1]

	if resourceQualifiers.IsSystemVariable(variableName) {
		return "", "", fmt.Errorf("system variable %q is not supported here, only user-defined variables are allowed", variableName)
	}

	scopedVariables, err := impl.scopedVariableService.GetScopedVariables(scope, []string{variableName}, true)
	if err != nil {
		impl.logger.Errorw("error resolving repo field variable", "variableName", variableName, "scope", scope, "err", err)
		return "", "", fmt.Errorf("error resolving variable %q: %w", variableName, err)
	}
	if len(scopedVariables) == 0 {
		return "", "", fmt.Errorf("variable %q not found, or has no value defined", variableName)
	}

	resolvedValue := scopedVariables[0].VariableValue.StringValue()
	return resolvedValue, variableName, nil
}

/*
 * Copyright (c) 2026. Devtron Inc.
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

package bean

import "testing"

func TestGitMaterialSetDefaultCloningMode(t *testing.T) {
	t.Run("defaults an omitted mode to full", func(t *testing.T) {
		material := &GitMaterial{}

		material.SetDefaultCloningMode()

		if material.CloningMode != GitMaterialCloningModeFull {
			t.Fatalf("expected %q, got %q", GitMaterialCloningModeFull, material.CloningMode)
		}
	})

	t.Run("preserves an explicitly selected mode", func(t *testing.T) {
		material := &GitMaterial{CloningMode: GitMaterialCloningModeShallow}

		material.SetDefaultCloningMode()

		if material.CloningMode != GitMaterialCloningModeShallow {
			t.Fatalf("expected %q, got %q", GitMaterialCloningModeShallow, material.CloningMode)
		}
	})
}

package overlay

import (
	"fmt"
	"strings"

	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/config"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/token"
	"gopkg.in/yaml.v3"
)

// ApplyTo will take an overlay and apply its changes to the given YAML
// document.
func (o *Overlay) ApplyTo(root *yaml.Node) error {
	// Priority is: remove > update > copy
	// Copy has no impact if remove is true or update contains a value
	for _, action := range o.Actions {
		var err error
		switch {
		case action.Remove:
			err = o.applyRemoveAction(root, action, nil)
		case !action.Update.IsZero():
			err = o.applyUpdateAction(root, action, &[]string{})
		case action.Copy != "":
			err = o.applyCopyAction(root, action, &[]string{})
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Overlay) ApplyToStrict(root *yaml.Node) ([]string, error) {
	multiError := []string{}
	warnings := []string{}
	hasFilterExpression := false

	// Priority is: remove > update > copy
	// Copy has no impact if remove is true or update contains a value
	for i, action := range o.Actions {
		tokens := token.NewTokenizer(action.Target, config.WithPropertyNameExtension()).Tokenize()
		for _, tok := range tokens {
			if tok.Token == token.FILTER {
				hasFilterExpression = true
			}
		}

		actionWarnings := []string{}
		err := o.validateSelectorHasAtLeastOneTarget(root, action)
		if err != nil {
			multiError = append(multiError, err.Error())
		}

		// Determine action type based on priority: remove > update > copy
		actionType := "unknown"
		switch {
		case action.Remove:
			actionType = "remove"
			err = o.applyRemoveAction(root, action, &actionWarnings)
		case !action.Update.IsZero():
			actionType = "update"
			err = o.applyUpdateAction(root, action, &actionWarnings)
		case action.Copy != "":
			actionType = "copy"
			err = o.applyCopyAction(root, action, &actionWarnings)
		default:
			err = fmt.Errorf("unknown action type: %v", action)
		}
		if err != nil {
			return nil, err
		}
		for _, warning := range actionWarnings {
			warnings = append(warnings, fmt.Sprintf("%s action (%v / %v) target=%s: %s", actionType, i+1, len(o.Actions), action.Target, warning))
		}
	}

	if hasFilterExpression && !o.UsesRFC9535() {
		warnings = append(warnings, "overlay has a filter expression but lacks `x-speakeasy-jsonpath: rfc9535` extension. Deprecated jsonpath behaviour in use. See overlay.speakeasy.com for the implementation playground.")
	}

	if len(multiError) > 0 {
		return warnings, fmt.Errorf("error applying overlay (strict): %v", strings.Join(multiError, ","))
	}
	return warnings, nil
}

func (o *Overlay) validateSelectorHasAtLeastOneTarget(root *yaml.Node, action Action) error {
	if action.Target == "" {
		return nil
	}

	p, err := o.NewPath(action.Target, nil)
	if err != nil {
		return err
	}

	nodes := p.Query(root)

	if len(nodes) == 0 {
		return fmt.Errorf("selector %q did not match any targets", action.Target)
	}

	// For copy actions, validate the source path (only if copy will actually be applied)
	// Copy has no impact if remove is true or update contains a value
	if action.Copy != "" && !action.Remove && action.Update.IsZero() {
		sourcePath, err := o.NewPath(action.Copy, nil)
		if err != nil {
			return err
		}

		sourceNodes := sourcePath.Query(root)
		if len(sourceNodes) == 0 {
			return fmt.Errorf("copy source selector %q did not match any nodes", action.Copy)
		}

		if len(sourceNodes) > 1 {
			return fmt.Errorf("copy source selector %q matched multiple nodes (%d), expected exactly one", action.Copy, len(sourceNodes))
		}
	}

	return nil
}

func (o *Overlay) applyRemoveAction(root *yaml.Node, action Action, warnings *[]string) error {
	if action.Target == "" {
		return nil
	}

	idx := newParentIndex(root)

	p, err := o.NewPath(action.Target, warnings)
	if err != nil {
		return err
	}

	nodes := p.Query(root)

	for _, node := range nodes {
		removeNode(idx, node)
	}

	return nil
}

func removeNode(idx parentIndex, node *yaml.Node) {
	parent := idx.getParent(node)
	if parent == nil {
		return
	}

	for i, child := range parent.Content {
		if child == node {
			switch parent.Kind {
			case yaml.MappingNode:
				if i%2 == 1 {
					// if we select a value, we should delete the key too
					parent.Content = append(parent.Content[:i-1], parent.Content[i+1:]...)
				} else {
					// if we select a key, we should delete the value
					parent.Content = append(parent.Content[:i], parent.Content[i+2:]...)
				}
				return
			case yaml.SequenceNode:
				parent.Content = append(parent.Content[:i], parent.Content[i+1:]...)
				return
			}
		}
	}
}

func (o *Overlay) applyUpdateAction(root *yaml.Node, action Action, warnings *[]string) error {
	if action.Target == "" {
		return nil
	}

	if action.Update.IsZero() {
		return nil
	}

	p, err := o.NewPath(action.Target, warnings)
	if err != nil {
		return err
	}

	nodes := p.Query(root)

	didMakeChange := false
	for _, node := range nodes {
		didMakeChange = updateNode(node, &action.Update) || didMakeChange
	}
	if !didMakeChange {
		*warnings = append(*warnings, "does nothing")
	}

	return nil
}

func updateNode(node *yaml.Node, updateNode *yaml.Node) bool {
	return mergeNode(node, updateNode)
}

func mergeNode(node *yaml.Node, merge *yaml.Node) bool {
	if node.Kind != merge.Kind {
		*node = *clone(merge)
		return true
	}
	switch node.Kind {
	default:
		isChanged := node.Value != merge.Value
		node.Value = merge.Value
		return isChanged
	case yaml.MappingNode:
		return mergeMappingNode(node, merge)
	case yaml.SequenceNode:
		return mergeSequenceNode(node, merge)
	}
}

// mergeMappingNode will perform a shallow merge of the merge node into the main
// node.
func mergeMappingNode(node *yaml.Node, merge *yaml.Node) bool {
	anyChange := false

	// If the target is an empty flow-style mapping and we're merging content,
	// convert to block style for better readability
	if len(node.Content) == 0 && node.Style == yaml.FlowStyle && len(merge.Content) > 0 {
		node.Style = 0 // Reset to default (block) style
	}

NextKey:
	for i := 0; i < len(merge.Content); i += 2 {
		mergeKey := merge.Content[i].Value
		mergeValue := merge.Content[i+1]

		for j := 0; j < len(node.Content); j += 2 {
			nodeKey := node.Content[j].Value
			if nodeKey == mergeKey {
				anyChange = mergeNode(node.Content[j+1], mergeValue) || anyChange
				continue NextKey
			}
		}

		node.Content = append(node.Content, merge.Content[i], clone(mergeValue))
		anyChange = true
	}
	return anyChange
}

// mergeSequenceNode will append the merge node's content to the original node.
func mergeSequenceNode(node *yaml.Node, merge *yaml.Node) bool {
	node.Content = append(node.Content, clone(merge).Content...)
	return true
}

func clone(node *yaml.Node) *yaml.Node {
	newNode := &yaml.Node{
		Kind:        node.Kind,
		Style:       node.Style,
		Tag:         node.Tag,
		Value:       node.Value,
		Anchor:      node.Anchor,
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
	}
	if node.Alias != nil {
		newNode.Alias = clone(node.Alias)
	}
	if node.Content != nil {
		newNode.Content = make([]*yaml.Node, len(node.Content))
		for i, child := range node.Content {
			newNode.Content[i] = clone(child)
		}
	}
	return newNode
}

// applyCopyAction applies a copy action to the document
// This is a stub implementation for the copy feature from Overlay Specification v1.1.0
func (o *Overlay) applyCopyAction(root *yaml.Node, action Action, warnings *[]string) error {
	if action.Target == "" {
		return nil
	}

	if action.Copy == "" {
		return nil
	}

	// Parse the source path
	sourcePath, err := o.NewPath(action.Copy, warnings)
	if err != nil {
		return fmt.Errorf("invalid copy source path %q: %w", action.Copy, err)
	}

	// Query the source nodes
	sourceNodes := sourcePath.Query(root)
	if len(sourceNodes) == 0 {
		// Source not found - in non-strict mode this is silently ignored
		// In strict mode, this will be caught by validateSelectorHasAtLeastOneTarget
		if warnings != nil {
			*warnings = append(*warnings, fmt.Sprintf("copy source %q not found", action.Copy))
		}
		return nil
	}

	if len(sourceNodes) > 1 {
		return fmt.Errorf("copy source path %q matched multiple nodes (%d), expected exactly one", action.Copy, len(sourceNodes))
	}

	sourceNode := sourceNodes[0]

	// Parse the target path
	targetPath, err := o.NewPath(action.Target, warnings)
	if err != nil {
		return fmt.Errorf("invalid target path %q: %w", action.Target, err)
	}

	// Query the target nodes
	targetNodes := targetPath.Query(root)

	// Copy the source node to each target
	didMakeChange := false
	for _, targetNode := range targetNodes {
		// Clone the source node to avoid reference issues
		copiedNode := clone(sourceNode)

		// Merge the copied node into the target
		didMakeChange = mergeNode(targetNode, copiedNode) || didMakeChange
	}

	if !didMakeChange && warnings != nil {
		*warnings = append(*warnings, "does nothing")
	}

	return nil
}

package yamlparser

// https://pkg.go.dev/gopkg.in/yaml.v3
// https://stackoverflow.com/questions/55674853/modify-existing-yaml-file-and-add-new-data-and-comments

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// TODO Test this.
//var data string = `
//---
//`

// Node Kind
//  1 = DocumentNode
//  2 = SequenceNode
//  4 = MappingNode
//  8 = ScalarNode
// 16 = AliasNode

func buildPrefix(column, insertAt int, insertChar string) string {
	var i int
	var char string
	builder := strings.Builder{}
	// Subtract 1 because the character *starts* at the column position.
	for i < column-1 {
		char = " "
		// Subtract 1 for the same reason.
		if insertAt > 0 && i == (insertAt-1) {
			char = insertChar
		}
		builder.WriteString(char)
		i += 1
	}
	return builder.String()
}

// TODO Review this.
//func buildStringNodes(key, value, comment string) []*yaml.Node {
//	keyNode := &yaml.Node{
//		Kind:        yaml.ScalarNode,
//		Tag:         "!!str",
//		Value:       key,
//		HeadComment: comment,
//	}
//	valueNode := &yaml.Node{
//		Kind:  yaml.ScalarNode,
//		Tag:   "!!str",
//		Value: value,
//	}
//	return []*yaml.Node{keyNode, valueNode}
//}

// TODO Review this.
func GetNode(node *yaml.Node, identifier string) *yaml.Node {
	returnNode := false
	for _, n := range node.Content {
		if n.Value == identifier {
			returnNode = true
			continue
		}
		if returnNode {
			return n
		}
		if len(n.Content) > 0 {
			ac_node := GetNode(n, identifier)
			if ac_node != nil {
				return ac_node
			}
		}
	}
	return nil
}

// TODO Add AliasNode support.
func WalkNodes(node *yaml.Node, prefix string) {
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			if node.HeadComment != "" {
				fmt.Println(node.HeadComment)
			}
			WalkNodes(node.Content[0], buildPrefix(node.Column, 0, ""))
			if node.FootComment != "" {
				fmt.Println(node.FootComment)
			}
		} else {
			fmt.Println("No Document node, nothing to do...")
			return
		}

	case yaml.MappingNode:
		if node.Tag == "!!map" && node.HeadComment != "" {
			fmt.Printf("%s%s\n", buildPrefix(node.Column-2, 0, ""), node.HeadComment)
		}
		nodes := node.Content
		var cur *yaml.Node
		var next *yaml.Node
		for i := 0; i < len(nodes); i += 1 {
			cur = nodes[i]
			if i+1 < len(nodes) {
				next = nodes[i+1]
				if next.Line == cur.Line {
					// TODO: Also add check for FootComment?
					if cur.HeadComment != "" {
						fmt.Printf("%s%s\n", buildPrefix(cur.Column, 0, ""), cur.HeadComment)
					}
					if next.Tag != "!!null" {
						WalkNodes(next, fmt.Sprintf("%s%s: ", prefix, cur.Value))
						// Do we need to check if the lines are the same for null tags?
					} else {
						cur.Value = cur.Value + ":"
						if cur.Column > 1 {
							WalkNodes(cur, buildPrefix(cur.Column, 1, "-"))
						} else {
							WalkNodes(cur, buildPrefix(cur.Column, 0, ""))
						}
					}
					// Reset the prefix in case we just came from the sequence
					// case which adds a hyphen (-) to the prefix.
					prefix = buildPrefix(cur.Column, 0, "")
					i += 1
				} else {
					cur.Value = cur.Value + ":"
					WalkNodes(cur, buildPrefix(cur.Column, 0, ""))
				}
			} else {
				WalkNodes(cur, buildPrefix(cur.Column, 0, "")+prefix)
			}
		}

	case yaml.ScalarNode:
		headComment := ""
		lineComment := ""
		footComment := ""

		if node.HeadComment != "" {
			if strings.Contains(node.HeadComment, "\n") {
				parts := strings.Split(node.HeadComment, "\n")
				node.HeadComment = strings.Join(parts, fmt.Sprintf("\n%s", buildPrefix(node.Column-2, 0, "")))
			}
			headComment = fmt.Sprintf("%s%s\n", buildPrefix(node.Column-2, 0, ""), node.HeadComment)
		}
		if node.LineComment != "" {
			lineComment = fmt.Sprintf(" %s", node.LineComment)
		}
		if node.FootComment != "" {
			// TODO Review this.
			if strings.Contains(node.FootComment, "\n") {
				parts := strings.Split(node.FootComment, "\n")
				node.FootComment = strings.Join(parts, fmt.Sprintf("\n%s", buildPrefix(node.Column-2, 0, "")))
			}
			footComment = fmt.Sprintf("\n%s%s", buildPrefix(node.Column-2, 0, ""), node.FootComment)
		}

		fmt.Printf("%s%s%v%s%s\n",
			headComment,
			prefix,
			node.Value,
			lineComment,
			footComment)

	case yaml.SequenceNode:
		for _, node := range node.Content {
			WalkNodes(node, buildPrefix(node.Column, node.Column-2, "-"))
		}
	}
}

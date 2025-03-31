package httputil

import (
	"strings"
)

// PathNode 表示 Trie 树中的一个节点
type PathNode struct {
	pattern  string      // 完整路由模式
	part     string      // 当前路径片段
	children []*PathNode // 子节点
	isWild   bool        // 是否为动态参数（如 :id 或 *filepath）
}

func (n *PathNode) Insert(pattern, split string) {
	parts := strings.Split(pattern, split)
	n.insert(pattern, parts, 0)
}

// matchChild 在当前节点的子节点中查找匹配的节点
func (n *PathNode) matchChild(part string) *PathNode {
	if n.children == nil {
		return nil
	}
	for _, child := range n.children {
		if child.part == part {
			return child
		}
	}
	return nil
}

func (n *PathNode) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}
	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		isWild := false
		if part != "" && (part[0] == ':' || part[0] == '*') {
			isWild = true
		}
		child = &PathNode{
			part:   part,
			isWild: isWild,
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *PathNode) Match(path, split string) (map[string]string, bool) {
	parts := strings.Split(path, split)
	retMap := make(map[string]string)
	pathNode := n.match(parts, &retMap, split, 0)
	if pathNode == nil {
		return map[string]string{}, false
	}
	return retMap, true
}

func (n *PathNode) match(parts []string, maps *map[string]string, split string, height int) *PathNode {
	if len(parts) == height || (n.part != "" && n.part[0] == '*') {
		if n.pattern != "" {
			return n
		}
		return nil
	}
	if n.children == nil {
		return nil
	}
	part := parts[height]
	for _, child := range n.children {
		if child.part == part || child.isWild { //匹配合适
			if child.isWild {
				childKey := child.part[1:]
				if child.part[0] == '*' {
					(*maps)[childKey] = strings.Join(parts[height:], split)
				} else {
					(*maps)[childKey] = part
				}
			}
			result := child.match(parts, maps, split, height+1)
			if result != nil {
				return result
			}
		}
	}
	return nil
}

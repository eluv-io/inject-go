package inject

import (
	"reflect"
	"sort"
	"strings"
)

// ctx is the dependency resolution context. It is used to detect circular
// dependencies and provide the dependency tree.
type ctx struct {
	root    *stack
	current *stack
}

func zeroCtx() ctx {
	return ctx{}
}

func newCtx(inj *injector) ctx {
	itype := "root"
	if inj.parent != nil {
		itype = "child"
	}
	root := newStack(nil, rootBindingKey{itype}, rootBinding{inj.name})
	return ctx{root: root, current: root}
}

func (c *ctx) push(key bindingKey, binding resolvedBinding) (err error) {
	if c.root == nil {
		return nil
	}
	c.current, err = c.current.push(key, binding)
	return err
}

func (c *ctx) pop() {
	if c.root == nil {
		return
	}
	c.current = c.current.pop()
}

func (c *ctx) tree() DependencyTree {
	if c.root == nil {
		return nil
	}
	dt := &dependencyTree{s: c.root}
	return dt
}

////////////////////////////////////////////////////////////////////////////////

type stack struct {
	parent   *stack
	children []*stack
	key      bindingKey
	binding  resolvedBinding
}

func newStack(parent *stack, key bindingKey, binding resolvedBinding) *stack {
	return &stack{
		parent:  parent,
		key:     key,
		binding: binding,
	}
}

func (s *stack) push(key bindingKey, binding resolvedBinding) (*stack, error) {
	// detect circular dependency
	for st := s; st != nil; st = st.parent {
		if st.binding == binding {
			// return s, so that caller can use result even if error
			return s,
				errCircularDependency.
					withTag("binding_key", " "+key.String(), true).
					withTag("call", " "+binding.String(), true).
					withTag("stack", "\n"+s.IndentString("\t* "))
		}
	}
	child := newStack(s, key, binding)
	s.children = append(s.children, child)
	return child, nil
}

func (s *stack) pop() *stack {
	if s.parent == nil {
		return s
	}
	return s.parent
}

func (s *stack) stack() []string {
	if s.parent == nil {
		return []string{s.binding.String()}
	}
	return append(s.parent.stack(), s.binding.String())
}

func (s *stack) String() string {
	return s.IndentString("")
}

func (s *stack) IndentString(ident string) string {
	sb := strings.Builder{}
	s.buildStack(&sb, ident)
	return sb.String()
}

func (s *stack) buildStack(sb *strings.Builder, ident string) {
	if s.parent != nil {
		s.parent.buildStack(sb, ident)
		sb.WriteString("\n")
	}
	sb.WriteString(ident)
	sb.WriteString(s.key.String())
	sb.WriteString(" : ")
	sb.WriteString(s.binding.String())
}

const (
	identReg    = "├── "
	identEnd    = "└── "
	identSubReg = "│   "
	identSubEnd = "    "
)

func (s *stack) print(sb *strings.Builder, ident, identSub string) {
	sb.WriteString(ident)
	sb.WriteString(s.key.String())
	sb.WriteString(" : ")
	sb.WriteString(s.binding.String())
	sb.WriteString("\n")

	sorted := make([]*stack, len(s.children))
	copy(sorted, s.children)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].String() < sorted[j].String()
	})
	last := len(s.children) - 1
	for idx, child := range sorted {
		if idx == last {
			child.print(sb, identSub+identEnd, identSub+identSubEnd)
		} else {
			child.print(sb, identSub+identReg, identSub+identSubReg)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type rootBindingKey struct {
	name string
}

func (r rootBindingKey) String() string {
	return r.name
}

func (r rootBindingKey) reflectType() reflect.Type {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type rootBinding struct {
	name string
}

func (r rootBinding) validate(c ctx) error {
	return nil
}

func (r rootBinding) get() (interface{}, error) {
	return nil, errIntermediateBinding
}

func (r rootBinding) String() string {
	return r.name
}

////////////////////////////////////////////////////////////////////////////////

type DependencyTree interface {
	String() string
}

type dependencyTree struct {
	s *stack
}

func (d *dependencyTree) String() string {
	sb := &strings.Builder{}
	d.s.print(sb, "", "")
	return sb.String()
}

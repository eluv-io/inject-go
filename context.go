package inject

import (
	"fmt"
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

func newCtx(inj *injector) ctx {
	itype := "root-injector"
	if inj.parent != nil {
		itype = "child-injector"
	}
	root := newStack(nil, rootBindingKey{itype}, rootBinding{inj.name})
	return ctx{root: root, current: root}
}

func (c *ctx) push(key bindingKey, binding resolvedBinding) (err error) {
	c.current, err = c.current.push(key, binding)
	return err
}

func (c *ctx) pop() {
	c.current = c.current.pop()
}

func (c *ctx) tree() DependencyTree {
	dt := &dependencyTree{s: c.root}
	return dt
}

////////////////////////////////////////////////////////////////////////////////

type stack struct {
	id       string
	parent   *stack
	children []*stack
	key      bindingKey
	binding  resolvedBinding
}

func newStack(parent *stack, key bindingKey, binding resolvedBinding) *stack {
	return &stack{
		id:      binding.String(),
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

func (s *stack) graphNodes(visited map[string]*stackRef) {
	id := s.binding.String()
	if sref, ok := visited[id]; ok {
		sref.incoming++
		return
	}
	visited[id] = &stackRef{s, 1}

	if len(s.children) == 0 {
		return
	}
	for _, child := range s.children {
		child.graphNodes(visited)
	}
}

func (s *stack) graphEdges(sb *strings.Builder) {
	if len(s.children) == 0 {
		return
	}
	sorted := make([]*stack, len(s.children))
	copy(sorted, s.children)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].String() < sorted[j].String()
	})
	for _, child := range sorted {
		sb.WriteString("\t")
		sb.WriteString(fmt.Sprintf("%q", s.binding.Short()))
		sb.WriteString(" -> ")
		sb.WriteString(fmt.Sprintf("%q", child.binding.Short()))
		sb.WriteString(";\n")
	}
}

func (s *stack) filter(accept func(bindingKey, resolvedBinding string) bool) *stack {
	cpy := &stack{
		id:      s.id,
		key:     s.key,
		binding: s.binding,
	}

	for _, child := range s.children {
		if accept(child.key.String(), child.binding.String()) {
			child = &stack{
				id:       child.id,
				parent:   cpy,
				children: child.children,
				key:      child.key,
				binding:  child.binding,
			}
			cpy.children = append(cpy.children, child)
			continue
		}
	}
	if len(cpy.children) > 0 || accept(s.key.String(), s.binding.String()) {
		return cpy
	}
	return nil
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

func (r rootBinding) Short() string {
	return r.name
}

////////////////////////////////////////////////////////////////////////////////

type DependencyTree interface {
	// String returns a string representation of the dependency tree.
	//
	// The dependency tree lists all bindings defined in the injector in the format "binding key : binding". The
	// children of a node are all dependencies of that binding, recursively. In other words: a node depends on all of
	// its children.
	//
	//	root-injector : app injector
	//	├── {type:*inject.Injector} : app injector
	//	├── {type:inject_test.A} : <github.com/eluv-io/inject-go_test.newA(inject_test.B, inject_test.E) inject_test.A>
	//	│   ├── {type:inject_test.B} : <github.com/eluv-io/inject-go_test.newB(inject_test.C, inject_test.D) inject_test.B>
	//	│   │   ├── {type:inject_test.C} : singleton inject_test.C
	//	│   │   └── {type:inject_test.D} : singleton inject_test.D
	//	│   └── {type:inject_test.E} : singleton inject_test.E
	//	├── {type:inject_test.B} : <github.com/eluv-io/inject-go_test.newB(inject_test.C, inject_test.D) inject_test.B>
	//	│   ├── {type:inject_test.C} : singleton inject_test.C
	//	│   └── {type:inject_test.D} : singleton inject_test.D
	//	├── {type:inject_test.C} : singleton inject_test.C
	//	├── {type:inject_test.D} : singleton inject_test.D
	//	└── {type:inject_test.E} : singleton inject_test.E
	String() string

	// Dot returns the dependency tree in DOT format, which can be used to visualize the dependency graph using tools
	// like Graphviz. The passed in options can be used to customize the graph and are inserted right before the node
	// and edge definitions.
	//
	// For example:
	//	layout=dot
	//	fontname="Helvetica,Arial,sans-serif"
	//	fontsize=10
	//	node [shape=box,width=0, height=0,fixedsize=false]
	//	edge [arrowsize=0.5]
	Dot(options ...string) string

	// Filter returns a new DependencyTree that only includes bindings that match the provided filter function.
	Filter(func(bindingKey, resolvedBinding string) bool) DependencyTree
}

type dependencyTree struct {
	s *stack
}

func (d *dependencyTree) String() string {
	if d == nil {
		return ""
	}
	sb := &strings.Builder{}
	d.s.print(sb, "", "")
	return sb.String()
}

func (d *dependencyTree) Dot(options ...string) string {
	if d == nil {
		return ""
	}
	sb := &strings.Builder{}
	nodes := make(map[string]*stackRef)

	d.s.graphNodes(nodes)

	sb.WriteString(`digraph G  {
`)
	for _, option := range options {
		sb.WriteString(option)
	}

	sorted := d.sort(nodes)
	for _, node := range sorted {
		sb.WriteString("\t")
		sb.WriteString(fmt.Sprintf("%q", node.binding.Short()))
		sb.WriteString(fmt.Sprintf(" [tooltip=%q]", node.binding.String()))
		sb.WriteString(";\n")
	}

	sb.WriteString("\n")

	for _, node := range sorted {
		node.graphEdges(sb)
	}

	sb.WriteString("}\n")

	return sb.String()
}

func (d *dependencyTree) Filter(accept func(bindingKey, resolvedBinding string) bool) DependencyTree {
	if d == nil {
		return nil
	}
	cpy := d.s.filter(accept)
	if cpy == nil {
		return nil
	}
	return &dependencyTree{cpy}
}

func (d *dependencyTree) sort(nodes map[string]*stackRef) []*stackRef {
	sorted := make([]*stackRef, 0, len(nodes))
	for _, node := range nodes {
		sorted = append(sorted, node)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].binding.Short() < sorted[j].binding.Short()
	})
	return sorted
}

////////////////////////////////////////////////////////////////////////////////

type stackRef struct {
	*stack
	incoming int
}

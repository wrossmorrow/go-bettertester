package bettertester

func dumbTest() {
	cg := NewCallGraph()
	cg.AddNewCallNamed("a")
	cg.AddNewCallNamed("b")
	cg.AddNewCallNamed("c")
	cg.AddNewCallNamed("d")
	cg.AddCallEdge("a", "b")
	cg.AddCallEdge("a", "c")
	cg.AddCallEdge("b", "c")
	cg.AddCallEdge("d", "b")

	cg.AddNewCallNamedWithRequestProto("e", &RequestProto{
		Scheme:     "http",
		Host:       "localhost",
		Port:       8080,
		Method:     "GET",
		Path:       "/",
		Headers:    make(map[string][]string),
		Parameters: make(map[string][]string),
		Body:       &BodyText{},
	})
	// cg.AddCallEdge("c", "e")
	// cg.AddCallEdge("e", "c")
	cg.AddCallEdge("c", "e")

	cg.Print()

	ctx := NewExecutionContext()
	err := cg.Execute(ctx)
	if err != nil {
		panic(err)
	}
}

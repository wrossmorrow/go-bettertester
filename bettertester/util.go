package bettertester

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func PrintIndented(d proto.Message) {
	opt := protojson.MarshalOptions{
		Indent: "  ",
	}
	b, _ := opt.Marshal(d)
	fmt.Printf("%v\n", string(b))
}

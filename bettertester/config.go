package bettertester

import (
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"

	pb "github.com/wrossmorrow/bettertester/gen/proto/go/config/v1"
)

func ReadConfig(filename string) (*pb.Config, error) {
	cfg := &pb.Config{}

	// read bytes from file
	y, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// convert yaml to json
	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return nil, err
	}

	// unmarshal json into config
	err = protojson.Unmarshal(j, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

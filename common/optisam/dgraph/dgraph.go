package dgraph

import (
	"github.com/pkg/errors"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

// NewDgraphConnection returns a new dgraph connection for the application.
func NewDgraphConnection(config *Config) (*dgo.Dgraph, error) {
	dgClients := make([]api.DgraphClient, 0, len(config.Hosts))
	for _, host := range config.Hosts {
		for i := 0; i < 2; i++ {
			conn, err := grpc.Dial(host, grpc.WithInsecure())
			if err != nil {
				return nil, errors.Wrap(err, "failed to open connection with dgraph")
			}
			dgClients = append(dgClients, api.NewDgraphClient(conn))
		}
	}

	return dgo.NewDgraphClient(dgClients...), nil
}

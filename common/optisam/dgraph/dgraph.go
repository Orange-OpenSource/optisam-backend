// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"github.com/pkg/errors"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"
)

// NewDgraphConnection returns a new dgraph connection for the application.
func NewDgraphConnection(config *Config) (*dgo.Dgraph, error) {
	dgClients := make([]api.DgraphClient, 0, len(config.Hosts))
	for _, host := range config.Hosts {
		conn, err := grpc.Dial(host, grpc.WithInsecure())
		if err != nil {
			return nil, errors.Wrap(err, "failed to open connection with dgraph")
		}
		dgClients = append(dgClients, api.NewDgraphClient(conn))
	}

	return dgo.NewDgraphClient(dgClients...), nil
}

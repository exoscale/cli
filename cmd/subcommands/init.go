/*
The purpose of this file is to initialize all of the commands, that are spread across different folders,
and to keep the imports in one place, keeping main.go clean.
*/

package subcommands

import (
	_ "github.com/exoscale/cli/cmd/compute/anti_affinity_group"
	_ "github.com/exoscale/cli/cmd/compute/blockstorage"
	_ "github.com/exoscale/cli/cmd/compute/deploy_target"
	_ "github.com/exoscale/cli/cmd/compute/elastic_ip"
	_ "github.com/exoscale/cli/cmd/compute/instance"
	_ "github.com/exoscale/cli/cmd/compute/instance_pool"
	_ "github.com/exoscale/cli/cmd/compute/instance_template"
	_ "github.com/exoscale/cli/cmd/compute/instance_type"
	_ "github.com/exoscale/cli/cmd/compute/load_balancer"
	_ "github.com/exoscale/cli/cmd/compute/private_network"
	_ "github.com/exoscale/cli/cmd/compute/security_group"
	_ "github.com/exoscale/cli/cmd/compute/sks"
	_ "github.com/exoscale/cli/cmd/compute/ssh_key"
	_ "github.com/exoscale/cli/cmd/dbaas"
	_ "github.com/exoscale/cli/cmd/iam"
)

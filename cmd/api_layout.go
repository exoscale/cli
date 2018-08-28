package cmd

import (
	"github.com/exoscale/egoscale"
)

type category struct {
	name  string
	alias []string
	doc   string
	cmd   []cmd
}

type cmd struct {
	command egoscale.Command
	name    string
	alias   []string
}

var methods = []category{
	{
		"network",
		[]string{"net"},
		"Network management",
		[]cmd{
			{&egoscale.CreateNetwork{}, "create", gCreateAlias},
			{&egoscale.DeleteNetwork{}, "delete", gDeleteAlias},
			{&egoscale.ListNetworkOfferings{}, "", nil},
			{&egoscale.ListNetworks{}, "list", gListAlias},
			{&egoscale.UpdateNetwork{}, "update", nil},
		},
	},
	{
		"vm",
		[]string{"virtual-machine"},
		"Virtual machine management",
		[]cmd{
			{&egoscale.AddNicToVirtualMachine{}, "addNic", nil},
			{&egoscale.ChangeServiceForVirtualMachine{}, "changeService", nil},
			{&egoscale.DeployVirtualMachine{}, "deploy", nil},
			{&egoscale.DestroyVirtualMachine{}, "destroy", nil},
			{&egoscale.ExpungeVirtualMachine{}, "expunge", nil},
			{&egoscale.GetVMPassword{}, "getPassword", nil},
			{&egoscale.GetVirtualMachineUserData{}, "getUserData", nil},
			{&egoscale.ListVirtualMachines{}, "list", gListAlias},
			{&egoscale.MigrateVirtualMachine{}, "", nil},
			{&egoscale.RebootVirtualMachine{}, "reboot", nil},
			{&egoscale.RecoverVirtualMachine{}, "recover", nil},
			{&egoscale.RemoveNicFromVirtualMachine{}, "removeNic", nil},
			{&egoscale.ResetPasswordForVirtualMachine{}, "resetPassword", nil},
			{&egoscale.RestoreVirtualMachine{}, "restore", nil},
			{&egoscale.ScaleVirtualMachine{}, "scale", nil},
			{&egoscale.StartVirtualMachine{}, "start", nil},
			{&egoscale.StopVirtualMachine{}, "stop", nil},
			{&egoscale.UpdateVirtualMachine{}, "update", nil},
			{&egoscale.UpdateVMAffinityGroup{}, "", nil},
			{&egoscale.DeleteReverseDNSFromVirtualMachine{}, "deleteReverseDNSFromVM", nil},
			{&egoscale.QueryReverseDNSForVirtualMachine{}, "queryReverseDNSForVM", nil},
			{&egoscale.UpdateReverseDNSForVirtualMachine{}, "updateReverseDNSForVM", nil},
		},
	},
	{
		"affinity-group",
		[]string{"ag"},
		"Affinity group management",
		[]cmd{
			{&egoscale.CreateAffinityGroup{}, "create", gCreateAlias},
			{&egoscale.DeleteAffinityGroup{}, "delete", gDeleteAlias},
			{&egoscale.ListAffinityGroups{}, "list", gListAlias},
		},
	},
	{
		"volume",
		[]string{"vol"},
		"Volume management",
		[]cmd{
			{&egoscale.ListVolumes{}, "list", gListAlias},
			{&egoscale.ResizeVolume{}, "resize", nil},
		},
	},
	{
		"template",
		[]string{"temp"},
		"Template management",
		[]cmd{
			{&egoscale.ListTemplates{}, "list", gListAlias},
		},
	},
	{
		"account",
		[]string{"acc"},
		"Account management",
		[]cmd{
			{&egoscale.ListAccounts{}, "list", gListAlias},
		},
	},
	{
		"zone",
		nil,
		"Zone management",
		[]cmd{
			{&egoscale.ListZones{}, "list", gListAlias},
		},
	},
	{
		"snapshot",
		[]string{"snap"},
		"Snapshot management",
		[]cmd{
			{&egoscale.CreateSnapshot{}, "create", gCreateAlias},
			{&egoscale.DeleteSnapshot{}, "delete", gDeleteAlias},
			{&egoscale.ListSnapshots{}, "list", gListAlias},
			{&egoscale.RevertSnapshot{}, "revert", nil},
		},
	},
	{
		"user",
		[]string{"usr"},
		"User management",
		[]cmd{
			{&egoscale.ListUsers{}, "list", gListAlias},
			{&egoscale.RegisterUserKeys{}, "", nil},
		},
	},
	{
		"security-group",
		[]string{"sg"},
		"Security group management",
		[]cmd{
			{&egoscale.AuthorizeSecurityGroupEgress{}, "authorizeEgress", nil},
			{&egoscale.AuthorizeSecurityGroupIngress{}, "authorizeIngress", nil},
			{&egoscale.CreateSecurityGroup{}, "create", gCreateAlias},
			{&egoscale.DeleteSecurityGroup{}, "delete", gDeleteAlias},
			{&egoscale.ListSecurityGroups{}, "list", gListAlias},
			{&egoscale.RevokeSecurityGroupEgress{}, "revokeEgress", nil},
			{&egoscale.RevokeSecurityGroupIngress{}, "revokeIngress", nil},
		},
	},
	{
		"ssh",
		nil,
		"SSH management",
		[]cmd{
			{&egoscale.RegisterSSHKeyPair{}, "register", nil},
			{&egoscale.ListSSHKeyPairs{}, "list", gListAlias},
			{&egoscale.CreateSSHKeyPair{}, "create", gCreateAlias},
			{&egoscale.DeleteSSHKeyPair{}, "delete", gDeleteAlias},
			{&egoscale.ResetSSHKeyForVirtualMachine{}, "reset", nil},
		},
	},
	{
		"vm-group",
		[]string{"vg"},
		"VM group management",
		[]cmd{
			{&egoscale.CreateInstanceGroup{}, "create", gCreateAlias},
			{&egoscale.ListInstanceGroups{}, "list", gListAlias},
		},
	},
	{
		"tag",
		nil,
		"Tags management",
		[]cmd{
			{&egoscale.CreateTags{}, "create", gCreateAlias},
			{&egoscale.DeleteTags{}, "delete", gDeleteAlias},
			{&egoscale.ListTags{}, "list", gListAlias},
		},
	},
	{
		"nic",
		nil,
		"Nic management",
		[]cmd{
			{&egoscale.ActivateIP6{}, "", nil},
			{&egoscale.AddIPToNic{}, "", nil},
			{&egoscale.ListNics{}, "list", gListAlias},
			{&egoscale.RemoveIPFromNic{}, "", nil},
		},
	},
	{
		"address",
		[]string{"addr"},
		"Address management",
		[]cmd{
			{&egoscale.AssociateIPAddress{}, "associate", gAssociateAlias},
			{&egoscale.DisassociateIPAddress{}, "disassociate", gDissociateAlias},
			{&egoscale.ListPublicIPAddresses{}, "list", gListAlias},
			{&egoscale.UpdateIPAddress{}, "update", nil},
			{&egoscale.DeleteReverseDNSFromPublicIPAddress{}, "deleteReverseDNSFromAddress", nil},
			{&egoscale.QueryReverseDNSForPublicIPAddress{}, "queryReverseDNSForAddress", nil},
			{&egoscale.UpdateReverseDNSForPublicIPAddress{}, "updateReverseDNSForAddress", nil},
		},
	},
	{
		"async-job",
		[]string{"aj"},
		"Async job management",
		[]cmd{
			{&egoscale.QueryAsyncJobResult{}, "", nil},
			{&egoscale.ListAsyncJobs{}, "", nil},
		},
	},
	{
		"api",
		nil,
		"Apis management",
		[]cmd{
			{&egoscale.ListAPIs{}, "list", gListAlias},
		},
	},
	{
		"event",
		nil,
		"Event management",
		[]cmd{
			{&egoscale.ListEventTypes{}, "listType", nil},
			{&egoscale.ListEvents{}, "list", gListAlias},
		},
	},
	{
		"offering",
		nil,
		"Offerings management",
		[]cmd{
			{&egoscale.ListResourceDetails{}, "listDetails", nil},
			{&egoscale.ListResourceLimits{}, "listLimits", nil},
			{&egoscale.ListServiceOfferings{}, "list", gListAlias},
		},
	},
}

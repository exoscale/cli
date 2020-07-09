package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/admin"
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
	hidden  bool
}

var t = true

var methods = []category{
	{
		"network",
		[]string{"net"},
		"Network management",
		[]cmd{
			{command: &egoscale.CreateNetwork{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.DeleteNetwork{}, name: "delete", alias: gDeleteAlias},
			{command: &egoscale.ListNetworks{}, name: "list", alias: gListAlias},
			{command: &egoscale.UpdateNetwork{}, name: "update"},
		},
	},
	{
		"vm",
		[]string{"virtual-machine"},
		"Virtual machine management",
		[]cmd{
			{command: &admin.ListVirtualMachines{ListAll: &t}, name: "listAll", hidden: true},
			{command: &egoscale.AddNicToVirtualMachine{}, name: "addNic"},
			{command: &egoscale.AttachISO{}},
			{command: &egoscale.ChangeServiceForVirtualMachine{}, name: "changeService"},
			{command: &egoscale.DeleteReverseDNSFromVirtualMachine{}, name: "deleteReverseDNSFromVM"},
			{command: &egoscale.DeployVirtualMachine{}, name: "deploy"},
			{command: &egoscale.DestroyVirtualMachine{}, name: "destroy"},
			{command: &egoscale.DetachISO{}},
			{command: &egoscale.ExpungeVirtualMachine{}, name: "expunge"},
			{command: &egoscale.GetVMPassword{}, name: "getPassword"},
			{command: &egoscale.GetVirtualMachineUserData{}, name: "getUserData"},
			{command: &egoscale.ListVirtualMachines{}, name: "list", alias: gListAlias},
			{command: &egoscale.QueryReverseDNSForVirtualMachine{}, name: "queryReverseDNSForVM"},
			{command: &egoscale.RebootVirtualMachine{}, name: "reboot"},
			{command: &egoscale.RecoverVirtualMachine{}, name: "recover"},
			{command: &egoscale.RemoveNicFromVirtualMachine{}, name: "removeNic"},
			{command: &egoscale.ResetPasswordForVirtualMachine{}, name: "resetPassword"},
			{command: &egoscale.RestoreVirtualMachine{}, name: "restore"},
			{command: &egoscale.ScaleVirtualMachine{}, name: "scale"},
			{command: &egoscale.StartVirtualMachine{}, name: "start"},
			{command: &egoscale.StopVirtualMachine{}, name: "stop"},
			{command: &egoscale.UpdateReverseDNSForVirtualMachine{}, name: "updateReverseDNSForVM"},
			{command: &egoscale.UpdateVMAffinityGroup{}, name: ""},
			{command: &egoscale.UpdateVMNicIP{}, name: "updateVMNicIP"},
			{command: &egoscale.UpdateVirtualMachine{}, name: "update"},
		},
	},
	{
		"affinity-group",
		[]string{"ag"},
		"Affinity group management",
		[]cmd{
			{command: &egoscale.CreateAffinityGroup{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.DeleteAffinityGroup{}, name: "delete", alias: gDeleteAlias},
			{command: &egoscale.ListAffinityGroups{}, name: "list", alias: gListAlias},
		},
	},
	{
		"volume",
		[]string{"vol"},
		"Volume management",
		[]cmd{
			{command: &egoscale.ListVolumes{}, name: "list", alias: gListAlias},
			{command: &egoscale.ResizeVolume{}, name: "resize"},
		},
	},
	{
		"template",
		[]string{"temp"},
		"Template management",
		[]cmd{
			{command: &egoscale.ListTemplates{}, name: "list", alias: gListAlias},
			{command: &egoscale.ListISOs{}},
		},
	},
	{
		"account",
		[]string{"acc"},
		"Account management",
		[]cmd{
			{command: &egoscale.ListAccounts{}, name: "list", alias: gListAlias},
		},
	},
	{
		"zone",
		nil,
		"Zone management",
		[]cmd{
			{command: &egoscale.ListZones{}, name: "list", alias: gListAlias},
		},
	},
	{
		"snapshot",
		[]string{"snap"},
		"Snapshot management",
		[]cmd{
			{command: &egoscale.CreateSnapshot{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.DeleteSnapshot{}, name: "delete", alias: gDeleteAlias},
			{command: &egoscale.ListSnapshots{}, name: "list", alias: gListAlias},
			{command: &egoscale.RevertSnapshot{}, name: "revert"},
			{command: &egoscale.ExportSnapshot{}, name: "export"},
		},
	},
	{
		"user",
		[]string{"usr"},
		"User management",
		[]cmd{
			{command: &egoscale.ListUsers{}, name: "list", alias: gListAlias},
			{command: &egoscale.RegisterUserKeys{}},
		},
	},
	{
		"security-group",
		[]string{"sg"},
		"Security group management",
		[]cmd{
			{command: &egoscale.AuthorizeSecurityGroupEgress{}, name: "authorizeEgress"},
			{command: &egoscale.AuthorizeSecurityGroupIngress{}, name: "authorizeIngress"},
			{command: &egoscale.CreateSecurityGroup{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.DeleteSecurityGroup{}, name: "delete", alias: gDeleteAlias},
			{command: &egoscale.ListSecurityGroups{}, name: "list", alias: gListAlias},
			{command: &egoscale.RevokeSecurityGroupEgress{}, name: "revokeEgress"},
			{command: &egoscale.RevokeSecurityGroupIngress{}, name: "revokeIngress"},
		},
	},
	{
		"ssh",
		nil,
		"SSH management",
		[]cmd{
			{command: &egoscale.RegisterSSHKeyPair{}, name: "register"},
			{command: &egoscale.ListSSHKeyPairs{}, name: "list", alias: gListAlias},
			{command: &egoscale.CreateSSHKeyPair{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.DeleteSSHKeyPair{}, name: "delete", alias: gDeleteAlias},
			{command: &egoscale.ResetSSHKeyForVirtualMachine{}, name: "reset"},
		},
	},
	{
		"vm-group",
		[]string{"vg"},
		"VM group management",
		[]cmd{
			{command: &egoscale.CreateInstanceGroup{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.DeleteInstanceGroup{}, name: "delete", alias: gDeleteAlias},
			{command: &egoscale.ListInstanceGroups{}, name: "list", alias: gListAlias},
			{command: &egoscale.UpdateInstanceGroup{}, name: "update"},
		},
	},
	{
		"tag",
		nil,
		"Tags management",
		[]cmd{
			{command: &egoscale.CreateTags{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.DeleteTags{}, name: "delete", alias: gDeleteAlias},
			{command: &egoscale.ListTags{}, name: "list", alias: gListAlias},
		},
	},
	{
		"nic",
		nil,
		"Nic management",
		[]cmd{
			{command: &egoscale.ActivateIP6{}},
			{command: &egoscale.AddIPToNic{}},
			{command: &egoscale.ListNics{}, name: "list", alias: gListAlias},
			{command: &egoscale.RemoveIPFromNic{}},
		},
	},
	{
		"address",
		[]string{"addr"},
		"Address management",
		[]cmd{
			{command: &egoscale.AssociateIPAddress{}, name: "associate", alias: gAssociateAlias},
			{command: &egoscale.DisassociateIPAddress{}, name: "disassociate", alias: gDissociateAlias},
			{command: &egoscale.ListPublicIPAddresses{}, name: "list", alias: gListAlias},
			{command: &egoscale.UpdateIPAddress{}, name: "update"},
			{command: &egoscale.DeleteReverseDNSFromPublicIPAddress{}, name: "deleteReverseDNSFromAddress"},
			{command: &egoscale.QueryReverseDNSForPublicIPAddress{}, name: "queryReverseDNSForAddress"},
			{command: &egoscale.UpdateReverseDNSForPublicIPAddress{}, name: "updateReverseDNSForAddress"},
		},
	},
	{
		"async-job",
		[]string{"aj"},
		"Async job management",
		[]cmd{
			{command: &egoscale.QueryAsyncJobResult{}},
			{command: &egoscale.ListAsyncJobs{}},
		},
	},
	{
		"api",
		nil,
		"Apis management",
		[]cmd{
			{command: &egoscale.ListAPIs{}, name: "list", alias: gListAlias},
		},
	},
	{
		"event",
		nil,
		"Event management",
		[]cmd{
			{command: &egoscale.ListEventTypes{}, name: "listType"},
			{command: &egoscale.ListEvents{}, name: "list", alias: gListAlias},
		},
	},
	{
		"offering",
		nil,
		"Offerings management",
		[]cmd{
			{command: &egoscale.ListResourceDetails{}, name: "listDetails"},
			{command: &egoscale.ListResourceLimits{}, name: "listLimits"},
			{command: &egoscale.ListServiceOfferings{}, name: "list", alias: gListAlias},
		},
	},
	{
		"instancepool",
		[]string{"ipool"},
		"Instance pool management",
		[]cmd{
			{command: &egoscale.CreateInstancePool{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.DestroyInstancePool{}, name: "destroy"},
			{command: &egoscale.ListInstancePools{}, name: "list", alias: gListAlias},
			{command: &egoscale.GetInstancePool{}, name: "get"},
			{command: &egoscale.UpdateInstancePool{}, name: "update"},
			{command: &egoscale.ScaleInstancePool{}, name: "scale"},
			{command: &egoscale.EvictInstancePoolMembers{}, name: "evict"},
		},
	},
	{
		"api-key",
		nil,
		"API keys management",
		[]cmd{
			{command: &egoscale.CreateAPIKey{}, name: "create", alias: gCreateAlias},
			{command: &egoscale.RevokeAPIKey{}, name: "revoke", alias: gRevokeAlias},
			{command: &egoscale.ListAPIKeys{}, name: "list", alias: gListAlias},
			{command: &egoscale.GetAPIKey{}, name: "get", alias: gListAlias},
		},
	},
}

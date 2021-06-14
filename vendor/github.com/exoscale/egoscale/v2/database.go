package v2

import (
	"context"
	"net/url"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// DatabaseBackupConfig represents a Database Backup configuration.
type DatabaseBackupConfig struct {
	FrequentIntervalMinutes    int64
	FrequentOldestAgeMinutes   int64
	InfrequentIntervalMinutes  int64
	InfrequentOldestAgeMinutes int64
	Interval                   int64
	MaxCount                   int64
	RecoveryMode               string
}

func databaseBackupConfigFromAPI(c *papi.DbaasBackupConfig) *DatabaseBackupConfig {
	return &DatabaseBackupConfig{
		FrequentIntervalMinutes:    papi.OptionalInt64(c.FrequentIntervalMinutes),
		FrequentOldestAgeMinutes:   papi.OptionalInt64(c.FrequentOldestAgeMinutes),
		InfrequentIntervalMinutes:  papi.OptionalInt64(c.InfrequentIntervalMinutes),
		InfrequentOldestAgeMinutes: papi.OptionalInt64(c.InfrequentOldestAgeMinutes),
		Interval:                   papi.OptionalInt64(c.Interval),
		MaxCount:                   papi.OptionalInt64(c.MaxCount),
		RecoveryMode:               papi.OptionalString(c.RecoveryMode),
	}
}

// DatabasePlan represents a Database Plan.
type DatabasePlan struct {
	BackupConfig     *DatabaseBackupConfig
	DiskSpace        int64
	MaxMemoryPercent int64
	Name             string
	Nodes            int64
	NodeCPUs         int64
	NodeMemory       int64
}

func databasePlanFromAPI(p *papi.DbaasPlan) *DatabasePlan {
	return &DatabasePlan{
		BackupConfig:     databaseBackupConfigFromAPI(p.BackupConfig),
		DiskSpace:        papi.OptionalInt64(p.DiskSpace),
		MaxMemoryPercent: papi.OptionalInt64(p.MaxMemoryPercent),
		Name:             *p.Name,
		Nodes:            papi.OptionalInt64(p.NodeCount),
		NodeCPUs:         papi.OptionalInt64(p.NodeCpuCount),
		NodeMemory:       papi.OptionalInt64(p.NodeMemory),
	}
}

// DatabaseServiceType represents a Database Service type.
type DatabaseServiceType struct {
	DefaultVersion   string
	Description      string
	LatestVersion    string
	Name             string
	Plans            []*DatabasePlan
	UserConfigSchema map[string]interface{}
}

func databaseServiceTypeFromAPI(t *papi.DbaasServiceType) *DatabaseServiceType {
	return &DatabaseServiceType{
		DefaultVersion: papi.OptionalString(t.DefaultVersion),
		Description:    papi.OptionalString(t.Description),
		LatestVersion:  papi.OptionalString(t.LatestVersion),
		Name:           string(*t.Name),
		Plans: func() []*DatabasePlan {
			plans := make([]*DatabasePlan, 0)
			if t.Plans != nil {
				for _, plan := range *t.Plans {
					plan := plan
					plans = append(plans, databasePlanFromAPI(&plan))
				}
			}
			return plans
		}(),
		UserConfigSchema: func() map[string]interface{} {
			if t.UserConfigSchema != nil {
				return t.UserConfigSchema.AdditionalProperties
			}
			return nil
		}(),
	}
}

// DatabaseServiceBackup represents a Database Service backup.
type DatabaseServiceBackup struct {
	Name string
	Size int64
	Date time.Time
}

// DatabaseServiceMaintenance represents a Database Service maintenance.
type DatabaseServiceMaintenance struct {
	DOW  string
	Time string
}

func databaseServiceMaintenanceFromAPI(m *papi.DbaasServiceMaintenance) *DatabaseServiceMaintenance {
	return &DatabaseServiceMaintenance{
		DOW:  string(m.Dow),
		Time: m.Time,
	}
}

func databaseServiceBackupFromAPI(b *papi.DbaasServiceBackup) *DatabaseServiceBackup {
	return &DatabaseServiceBackup{
		Name: b.BackupName,
		Size: b.DataSize,
		Date: b.BackupTime,
	}
}

// DatabaseServiceUser represents a Database Service user.
type DatabaseServiceUser struct {
	Password string
	Type     string
	UserName string
}

func databaseServiceUserFromAPI(u *papi.DbaasServiceUser) *DatabaseServiceUser {
	return &DatabaseServiceUser{
		Password: papi.OptionalString(u.Password),
		UserName: u.Username,
		Type:     u.Type,
	}
}

// DatabaseService represents a Database Service.
type DatabaseService struct {
	Backups               []*DatabaseServiceBackup
	ConnectionInfo        map[string]interface{}
	CreatedAt             time.Time
	Description           string
	DiskSize              int64
	Features              map[string]interface{}
	Maintenance           *DatabaseServiceMaintenance
	Metadata              map[string]interface{}
	Name                  string
	Nodes                 int64
	NodeCPUs              int64
	NodeMemory            int64
	Plan                  string
	State                 string
	TerminationProtection bool
	Type                  string
	UpdatedAt             time.Time
	URI                   *url.URL
	UserConfig            map[string]interface{}
	Users                 []*DatabaseServiceUser
}

func databaseServiceFromAPI(s *papi.DbaasService) *DatabaseService {
	return &DatabaseService{
		Backups: func() []*DatabaseServiceBackup {
			backups := make([]*DatabaseServiceBackup, 0)
			if s.Backups != nil {
				for _, b := range *s.Backups {
					backups = append(backups, databaseServiceBackupFromAPI(&b))
				}
			}
			return backups
		}(),
		ConnectionInfo: func() (v map[string]interface{}) {
			if s.ConnectionInfo != nil {
				v = s.ConnectionInfo.AdditionalProperties
			}
			return
		}(),
		CreatedAt: func() (v time.Time) {
			if s.CreatedAt != nil {
				v = *s.CreatedAt
			}
			return v
		}(),
		Description: papi.OptionalString(s.Description),
		DiskSize:    papi.OptionalInt64(s.DiskSize),
		Features: func() (v map[string]interface{}) {
			if s.Features != nil {
				v = s.Features.AdditionalProperties
			}
			return
		}(),
		Maintenance: func() (v *DatabaseServiceMaintenance) {
			if s.Maintenance != nil {
				return databaseServiceMaintenanceFromAPI(s.Maintenance)
			}
			return
		}(),
		Metadata: func() (v map[string]interface{}) {
			if s.Metadata != nil {
				v = s.Metadata.AdditionalProperties
			}
			return
		}(),
		Name:       string(s.Name),
		Nodes:      papi.OptionalInt64(s.NodeCount),
		NodeCPUs:   papi.OptionalInt64(s.NodeCpuCount),
		NodeMemory: papi.OptionalInt64(s.NodeMemory),
		Plan:       s.Plan,
		State: func() (v string) {
			if s.State != nil {
				v = string(*s.State)
			}
			return v
		}(),
		TerminationProtection: papi.OptionalBool(s.TerminationProtection),
		Type:                  string(s.Type),
		UpdatedAt: func() (v time.Time) {
			if s.UpdatedAt != nil {
				v = *s.UpdatedAt
			}
			return v
		}(),
		URI: func() *url.URL {
			if s.Uri != nil {
				if u, _ := url.Parse(*s.Uri); u != nil {
					return u
				}
			}
			return nil
		}(),
		UserConfig: func() (v map[string]interface{}) {
			if s.UserConfig != nil {
				v = s.UserConfig.AdditionalProperties
			}
			return
		}(),
		Users: func() []*DatabaseServiceUser {
			users := make([]*DatabaseServiceUser, 0)
			if s.Users != nil {
				for _, u := range *s.Users {
					users = append(users, databaseServiceUserFromAPI(&u))
				}
			}
			return users
		}(),
	}
}

// ListDatabaseServiceTypes returns the list of existing Database Service types.
func (c *Client) ListDatabaseServiceTypes(ctx context.Context, zone string) ([]*DatabaseServiceType, error) {
	list := make([]*DatabaseServiceType, 0)

	resp, err := c.ListDbaasServiceTypesWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.DbaasServiceTypes != nil {
		for i := range *resp.JSON200.DbaasServiceTypes {
			list = append(list, databaseServiceTypeFromAPI(&(*resp.JSON200.DbaasServiceTypes)[i]))
		}
	}

	return list, nil
}

// GetDatabaseServiceType returns the Database Service type corresponding to the specified name.
func (c *Client) GetDatabaseServiceType(ctx context.Context, zone, name string) (*DatabaseServiceType, error) {
	resp, err := c.GetDbaasServiceTypeWithResponse(apiv2.WithZone(ctx, zone), name)
	if err != nil {
		return nil, err
	}

	return databaseServiceTypeFromAPI(resp.JSON200), nil
}

// CreateDatabaseService creates a Database Service.
func (c *Client) CreateDatabaseService(
	ctx context.Context,
	zone string,
	databaseService *DatabaseService,
) (*DatabaseService, error) {
	_, err := c.CreateDbaasServiceWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreateDbaasServiceJSONRequestBody{
			Maintenance: func() (v *struct {
				Dow  papi.CreateDbaasServiceJSONBodyMaintenanceDow `json:"dow"`
				Time string                                        `json:"time"`
			}) {
				if databaseService.Maintenance != nil {
					v = &struct {
						Dow  papi.CreateDbaasServiceJSONBodyMaintenanceDow `json:"dow"`
						Time string                                        `json:"time"`
					}{
						Dow:  papi.CreateDbaasServiceJSONBodyMaintenanceDow(databaseService.Maintenance.DOW),
						Time: databaseService.Maintenance.Time,
					}
				}
				return
			}(),
			Name:                  papi.DbaasServiceName(databaseService.Name),
			Plan:                  databaseService.Plan,
			TerminationProtection: &databaseService.TerminationProtection,
			Type:                  papi.DbaasServiceTypeName(databaseService.Type),
			UserConfig: func() *papi.CreateDbaasServiceJSONBody_UserConfig {
				if len(databaseService.UserConfig) > 0 {
					return &papi.CreateDbaasServiceJSONBody_UserConfig{
						AdditionalProperties: databaseService.UserConfig,
					}
				}
				return nil
			}(),
		})
	if err != nil {
		return nil, err
	}

	return c.GetDatabaseService(ctx, zone, databaseService.Name)
}

// ListDatabaseServices returns the list of Database Services.
func (c *Client) ListDatabaseServices(ctx context.Context, zone string) ([]*DatabaseService, error) {
	list := make([]*DatabaseService, 0)

	resp, err := c.ListDbaasServicesWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.DbaasServices != nil {
		for i := range *resp.JSON200.DbaasServices {
			list = append(list, databaseServiceFromAPI(&(*resp.JSON200.DbaasServices)[i]))
		}
	}

	return list, nil
}

// GetDatabaseService returns the Database Service corresponding to the specified name.
func (c *Client) GetDatabaseService(ctx context.Context, zone, name string) (*DatabaseService, error) {
	resp, err := c.GetDbaasServiceWithResponse(apiv2.WithZone(ctx, zone), name)
	if err != nil {
		return nil, err
	}

	return databaseServiceFromAPI(resp.JSON200), nil
}

// UpdateDatabaseService updates the specified Database Service.
func (c *Client) UpdateDatabaseService(ctx context.Context, zone string, databaseService *DatabaseService) error {
	_, err := c.UpdateDbaasServiceWithResponse(
		apiv2.WithZone(ctx, zone),
		databaseService.Name,
		papi.UpdateDbaasServiceJSONRequestBody{
			Maintenance: func() (v *struct {
				Dow  papi.UpdateDbaasServiceJSONBodyMaintenanceDow `json:"dow"`
				Time string                                        `json:"time"`
			}) {
				if databaseService.Maintenance != nil {
					v = &struct {
						Dow  papi.UpdateDbaasServiceJSONBodyMaintenanceDow `json:"dow"`
						Time string                                        `json:"time"`
					}{
						Dow:  papi.UpdateDbaasServiceJSONBodyMaintenanceDow(databaseService.Maintenance.DOW),
						Time: databaseService.Maintenance.Time,
					}
				}
				return
			}(),
			Plan: func() (v *string) {
				if databaseService.Plan != "" {
					v = &databaseService.Plan
				}
				return
			}(),
			TerminationProtection: &databaseService.TerminationProtection,
			UserConfig: func() *papi.UpdateDbaasServiceJSONBody_UserConfig {
				if len(databaseService.UserConfig) > 0 {
					return &papi.UpdateDbaasServiceJSONBody_UserConfig{
						AdditionalProperties: databaseService.UserConfig,
					}
				}
				return nil
			}(),
		})
	if err != nil {
		return err
	}

	return nil
}

// DeleteDatabaseService deletes the specified Database Service.
func (c *Client) DeleteDatabaseService(ctx context.Context, zone, name string) error {
	_, err := c.TerminateDbaasServiceWithResponse(apiv2.WithZone(ctx, zone), name)
	if err != nil {
		return err
	}

	return nil
}

package mcs

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type dbClusterStatus string

var (
	dbClusterStatusActive   dbClusterStatus = "CLUSTER_ACTIVE"
	dbClusterStatusBuild    dbClusterStatus = "BUILDING"
	dbClusterStatusDeleted  dbClusterStatus = "DELETED"
	dbClusterStatusDeleting dbClusterStatus = "DELETING"
	dbClusterStatusGrow     dbClusterStatus = "GROWING_CLUSTER"
	dbClusterStatusResize   dbClusterStatus = "RESIZING_CLUSTER"
	dbClusterStatusShrink   dbClusterStatus = "SHRINKING_CLUSTER"
	dbClusterStatusUpdating dbClusterStatus = "UPDATING_CLUSTER"
)

func resourceDatabaseCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseClusterCreate,
		Read:   resourceDatabaseClusterRead,
		Delete: resourceDatabaseClusterDelete,
		Update: resourceDatabaseClusterUpdate,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(dbCreateTimeout),
			Delete: schema.DefaultTimeout(dbDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
				Computed: false,
			},

			"cluster_size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},

			"volume_size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
				Computed: false,
			},

			"volume_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
				Computed: false,
			},

			"wal_volume": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: false,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"autoexpand": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"max_disk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},

			"datastore": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if v != Galera && v != Postgres {
									errs = append(errs, fmt.Errorf("datastore type must be one of %v, got: %s", getClusterDatastores(), v))
								}
								return
							},
						},
					},
				},
			},

			"network": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"port": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"fixed_ip_v4": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"configuration_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
				ForceNew: false,
			},

			"root_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},

			"root_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Computed:  true,
				ForceNew:  false,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
				ForceNew: true,
			},

			"floating_ip_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: false,
				ForceNew: true,
			},

			"keypair": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
				ForceNew: true,
			},

			"disk_autoexpand": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"autoexpand": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"max_disk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},

			"capabilities": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"settings": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
		},
	}
}

func resourceDatabaseClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	createOpts := &dbClusterCreateOpts{
		Name:              d.Get("name").(string),
		FloatingIPEnabled: d.Get("floating_ip_enabled").(bool),
	}

	message := "unable to determine mcs_db_instance"
	if v, ok := d.GetOk("datastore"); ok {
		datastore, err := extractDatabaseInstanceDatastore(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s datastore", message)
		}
		createOpts.Datastore = &datastore
	}

	if v, ok := d.GetOk("disk_autoexpand"); ok {
		autoExpandOpts, err := extractDatabaseInstanceAutoExpand(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s disk_autoexpand", message)
		}
		if autoExpandOpts.AutoExpand {
			createOpts.AutoExpand = 1
		} else {
			createOpts.AutoExpand = 0
		}
		createOpts.MaxDiskSize = autoExpandOpts.MaxDiskSize
	}

	clusterSize := d.Get("cluster_size").(int)
	instances := make([]dbClusterInstanceCreateOpts, clusterSize)
	volumeSize := d.Get("volume_size").(int)
	createDBInstanceOpts := dbClusterInstanceCreateOpts{
		Keypair:          d.Get("keypair").(string),
		AvailabilityZone: d.Get("availability_zone").(string),
		FlavorRef:        d.Get("flavor_id").(string),
		Volume:           &volume{Size: &volumeSize, VolumeType: d.Get("volume_type").(string)},
	}

	if v, ok := d.GetOk("network"); ok {
		createDBInstanceOpts.Nics, err = extractDatabaseInstanceNetworks(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s network", message)
		}
	}
	if capabilities, ok := d.GetOk("capabilities"); ok {
		capabilitiesOpts, err := extractDatabaseInstanceCapabilities(capabilities.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s capability", message)
		}
		createDBInstanceOpts.Capabilities = capabilitiesOpts
	}

	if v, ok := d.GetOk("wal_volume"); ok {
		walVolumeOpts, err := extractDatabaseInstanceWalVolume(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s wal_volume", message)
		}
		createDBInstanceOpts.Walvolume = &walVolume{
			Size:        &walVolumeOpts.Size,
			VolumeType:  walVolumeOpts.VolumeType,
			MaxDiskSize: walVolumeOpts.MaxDiskSize,
		}
		if walVolumeOpts.AutoExpand {
			createDBInstanceOpts.Walvolume.AutoExpand = 1
		} else {
			createDBInstanceOpts.Walvolume.AutoExpand = 0
		}
	}

	for i := 0; i < clusterSize; i++ {
		instances[i] = createDBInstanceOpts
	}

	createOpts.Instances = instances

	log.Printf("[DEBUG] mcs_db_cluster create options: %#v", createOpts)
	clust := dbCluster{}
	clust.Cluster = createOpts

	cluster, err := dbClusterCreate(DatabaseV1Client, clust).extract()
	if err != nil {
		return fmt.Errorf("error creating mcs_db_instance: %s", err)
	}

	// Wait for the cluster to become available.
	log.Printf("[DEBUG] Waiting for mcs_db_cluster %s to become available", cluster.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{string(dbClusterStatusBuild)},
		Target:     []string{string(dbClusterStatusActive)},
		Refresh:    databaseClusterStateRefreshFunc(DatabaseV1Client, cluster.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      dbInstanceDelay,
		MinTimeout: dbInstanceMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_db_cluster %s to become ready: %s", cluster.ID, err)
	}

	if configuration, ok := d.GetOk("configuration_id"); ok {
		log.Printf("[DEBUG] Attaching configuration %s to mcs_db_cluster %s", configuration, cluster.ID)
		var attachConfigurationOpts dbClusterAttachConfigurationGroupOpts
		attachConfigurationOpts.ConfigurationAttach.ConfigurationID = configuration.(string)
		err := instanceAttachConfigurationGroup(DatabaseV1Client, cluster.ID, &attachConfigurationOpts).ExtractErr()
		if err != nil {
			return fmt.Errorf("error attaching configuration group %s to mcs_db_instance %s: %s",
				configuration, cluster.ID, err)
		}
	}

	// Store the ID now
	d.SetId(cluster.ID)
	return resourceDatabaseClusterRead(d, meta)
}

func resourceDatabaseClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	cluster, err := dbClusterGet(DatabaseV1Client, d.Id()).extract()
	if err != nil {
		return checkDeleted(d, err, "Error retrieving mcs_db_cluster")
	}

	log.Printf("[DEBUG] Retrieved mcs_db_cluster %s: %#v", d.Id(), cluster)

	d.Set("name", cluster.Name)
	d.Set("datastore", cluster.DataStore)

	return nil
}

func resourceDatabaseClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{string(dbClusterStatusBuild)},
		Target:     []string{string(dbClusterStatusActive)},
		Refresh:    databaseClusterStateRefreshFunc(DatabaseV1Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      dbInstanceDelay,
		MinTimeout: dbInstanceMinTimeout,
	}

	if d.HasChange("configuration_id") {
		old, new := d.GetChange("configuration_id")

		var detachConfigurationOpts dbClusterDetachConfigurationGroupOpts
		detachConfigurationOpts.ConfigurationDetach.ConfigurationID = old.(string)
		err := dbClusterAction(DatabaseV1Client, d.Id(), &detachConfigurationOpts).ExtractErr()
		if err != nil {
			return err
		}
		log.Printf("Detaching configuration %s from mcs_db_cluster %s", old, d.Id())

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_cluster %s to become ready: %s", d.Id(), err)
		}

		if new != "" {
			var attachConfigurationOpts dbClusterAttachConfigurationGroupOpts
			attachConfigurationOpts.ConfigurationAttach.ConfigurationID = new.(string)
			err := dbClusterAction(DatabaseV1Client, d.Id(), &attachConfigurationOpts).ExtractErr()
			if err != nil {
				return err
			}
			log.Printf("Attaching configuration %s to mcs_db_cluster %s", new, d.Id())

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("error waiting for mcs_db_cluster %s to become ready: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("volume_size") {
		old, new := d.GetChange("volume_size")
		if new.(int) < old.(int) {
			return fmt.Errorf("the new volume size %d must be larger than the current volume size of %d", new.(int), old.(int))
		}
		var resizeVolumeOpts dbClusterResizeVolumeOpts
		resizeVolumeOpts.Resize.Volume.Size = new.(int)
		err := dbClusterAction(DatabaseV1Client, d.Id(), &resizeVolumeOpts).ExtractErr()
		if err != nil {
			return err
		}
		log.Printf("Resizing volume from mcs_db_cluster %s", d.Id())

		stateConf.Pending = []string{string(dbClusterStatusResize)}
		stateConf.Target = []string{string(dbClusterStatusActive)}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_cluster %s to become ready: %s", d.Id(), err)
		}
	}

	if d.HasChange("flavor_id") {
		var resizeOpts dbClusterResizeOpts
		resizeOpts.Resize.FlavorRef = d.Get("flavor_id").(string)
		err := dbClusterAction(DatabaseV1Client, d.Id(), &resizeOpts).ExtractErr()
		if err != nil {
			return err
		}
		log.Printf("Resizing flavor from mcs_db_cluster %s", d.Id())

		stateConf.Pending = []string{string(dbClusterStatusResize)}
		stateConf.Target = []string{string(dbClusterStatusActive)}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_cluster %s to become ready: %s", d.Id(), err)
		}
	}

	if d.HasChange("disk_autoexpand") {
		_, new := d.GetChange("disk_autoexpand")
		autoExpandProperties, err := extractDatabaseInstanceAutoExpand(new.([]interface{}))
		if err != nil {
			return fmt.Errorf("unable to determine mcs_db_cluster disk_autoexpand")
		}
		var autoExpandOpts dbClusterUpdateAutoExpandOpts
		if autoExpandProperties.AutoExpand {
			autoExpandOpts.Cluster.VolumeAutoresizeEnabled = 1
		} else {
			autoExpandOpts.Cluster.VolumeAutoresizeEnabled = 0
		}
		autoExpandOpts.Cluster.VolumeAutoresizeMaxSize = autoExpandProperties.MaxDiskSize
		err = dbClusterUpdateAutoExpand(DatabaseV1Client, d.Id(), &autoExpandOpts).ExtractErr()
		if err != nil {
			return err
		}

		stateConf.Pending = []string{string(dbClusterStatusUpdating)}
		stateConf.Target = []string{string(dbClusterStatusActive)}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_cluster %s to become ready: %s", d.Id(), err)
		}
	}

	if d.HasChange("wal_volume") {
		old, new := d.GetChange("wal_volume")
		walVolumeOptsNew, err := extractDatabaseInstanceWalVolume(new.([]interface{}))
		if err != nil {
			return fmt.Errorf("unable to determine mcs_db_instance wal_volume")
		}

		walVolumeOptsOld, err := extractDatabaseInstanceWalVolume(old.([]interface{}))
		if err != nil {
			return fmt.Errorf("unable to determine mcs_db_instance wal_volume")
		}

		if walVolumeOptsNew.Size != walVolumeOptsOld.Size {
			if walVolumeOptsNew.Size < walVolumeOptsOld.Size {
				return fmt.Errorf("the new wal volume size %d must be larger than the current volume size of %d", walVolumeOptsNew.Size, walVolumeOptsOld.Size)
			}
			var resizeWalVolumeOpts dbClusterResizeWalVolumeOpts
			resizeWalVolumeOpts.Resize.Volume.Size = walVolumeOptsNew.Size
			resizeWalVolumeOpts.Resize.Volume.Kind = "wal"
			err = dbClusterAction(DatabaseV1Client, d.Id(), &resizeWalVolumeOpts).ExtractErr()
			if err != nil {
				return err
			}

			stateConf.Pending = []string{string(dbClusterStatusResize)}
			stateConf.Target = []string{string(dbClusterStatusActive)}

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", d.Id(), err)
			}
		}

		// Wal volume autoresize params update
		var autoExpandWalOpts dbClusterUpdateAutoExpandWalOpts
		if walVolumeOptsNew.AutoExpand {
			autoExpandWalOpts.Cluster.WalVolume.VolumeAutoresizeEnabled = 1
		} else {
			autoExpandWalOpts.Cluster.WalVolume.VolumeAutoresizeEnabled = 0
		}
		autoExpandWalOpts.Cluster.WalVolume.VolumeAutoresizeMaxSize = walVolumeOptsNew.MaxDiskSize
		err = dbClusterUpdateAutoExpand(DatabaseV1Client, d.Id(), &autoExpandWalOpts).ExtractErr()
		if err != nil {
			return err
		}
	}

	if d.HasChange("capabilities") {
		_, newCapabilities := d.GetChange("capabilities")
		newCapabilitiesOpts, err := extractDatabaseInstanceCapabilities(newCapabilities.([]interface{}))
		if err != nil {
			return fmt.Errorf("unable to determine mcs_db_instance capability")
		}
		var applyCapabilityOpts dbClusterApplyCapabilityOpts
		applyCapabilityOpts.ApplyCapability.Capabilities = newCapabilitiesOpts

		err = dbClusterAction(DatabaseV1Client, d.Id(), &applyCapabilityOpts).ExtractErr()

		if err != nil {
			return fmt.Errorf("error applying capability to mcs_db_instance %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("cluster_size") {
		old, new := d.GetChange("cluster_size")
		if new.(int) > old.(int) {
			opts := make([]dbClusterGrowOpts, new.(int)-old.(int))

			volumeSize := d.Get("volume_size").(int)
			growOpts := dbClusterGrowOpts{
				Keypair:          d.Get("keypair").(string),
				AvailabilityZone: d.Get("availability_zone").(string),
				FlavorRef:        d.Get("flavor_id").(string),
				Volume:           &volume{Size: &volumeSize, VolumeType: d.Get("volume_type").(string)},
			}
			if v, ok := d.GetOk("wal_volume"); ok {
				walVolumeOpts, err := extractDatabaseInstanceWalVolume(v.([]interface{}))
				if err != nil {
					return fmt.Errorf("unable to determine mcs_db_cluster wal_volume")
				}
				growOpts.Walvolume = &walVolume{
					Size:       &walVolumeOpts.Size,
					VolumeType: walVolumeOpts.VolumeType,
				}
			}
			for i := 0; i < len(opts); i++ {
				opts[i] = growOpts
			}
			growClusterOpts := dbClusterGrowClusterOpts{
				Grow: opts,
			}
			err = dbClusterAction(DatabaseV1Client, d.Id(), &growClusterOpts).ExtractErr()

			if err != nil {
				return fmt.Errorf("error growing mcs_db_cluster %s: %s", d.Id(), err)
			}
			stateConf.Pending = []string{string(dbClusterStatusGrow)}
			stateConf.Target = []string{string(dbClusterStatusActive)}

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("error waiting for mcs_db_cluster %s to become ready: %s", d.Id(), err)
			}
		} else {
			cluster, err := dbClusterGet(DatabaseV1Client, d.Id()).extract()
			if err != nil {
				return checkDeleted(d, err, "Error retrieving mcs_db_cluster")
			}
			ids := make([]dbClusterShrinkOpts, old.(int)-new.(int))
			for i := 0; i < len(ids); i++ {
				ids[i].ID = cluster.Instances[i].ID
			}

			shrinkClusterOpts := dbClusterShrinkClusterOpts{
				Shrink: ids,
			}

			err = dbClusterAction(DatabaseV1Client, d.Id(), &shrinkClusterOpts).ExtractErr()

			if err != nil {
				return fmt.Errorf("error growing mcs_db_cluster %s: %s", d.Id(), err)
			}
			stateConf.Pending = []string{string(dbClusterStatusShrink)}
			stateConf.Target = []string{string(dbClusterStatusActive)}

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("error waiting for mcs_db_cluster %s to become ready: %s", d.Id(), err)
			}
		}
	}

	return resourceDatabaseClusterRead(d, meta)
}

func resourceDatabaseClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	err = clusterDelete(DatabaseV1Client, d.Id()).ExtractErr()
	if err != nil {
		return checkDeleted(d, err, "Error deleting mcs_db_cluster")
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{string(dbClusterStatusActive), string(dbClusterStatusDeleting)},
		Target:     []string{string(dbClusterStatusDeleted)},
		Refresh:    databaseClusterStateRefreshFunc(DatabaseV1Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      dbInstanceDelay,
		MinTimeout: dbInstanceMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_db_cluster %s to delete: %s", d.Id(), err)
	}

	return nil
}

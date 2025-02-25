package mcs

func baseURL(c ContainerClient, api string) string {
	return c.ServiceURL(api)
}

func getURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id)
}

func deleteURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id)
}

func kubeConfigURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "kube_config")
}

func actionsURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "actions")
}

func upgradeURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "actions", "upgrade")
}

func scaleURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "actions", "scale")
}

func instanceActionURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "action")
}

func rootUserURL(c databaseClient, api string, id string) string {
	return c.ServiceURL(api, id, "root")
}

func userDatabasesURL(c databaseClient, api string, id string, userName string) string {
	return c.ServiceURL(api, id, "users", userName, "databases")
}

func userDatabaseURL(c databaseClient, api string, id string, userName string, databaseName string) string {
	return c.ServiceURL(api, id, "users", userName, "databases", databaseName)
}

func userURL(c databaseClient, api string, id string, userName string) string {
	return c.ServiceURL(api, id, "users", userName)
}

func instanceDatabasesURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "databases")
}

func instanceUsersURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "users")
}

func instanceUserURL(c ContainerClient, api string, id string, userName string) string {
	return c.ServiceURL(api, id, "users", userName)
}

func instanceDatabaseURL(c ContainerClient, api string, id string, databaseName string) string {
	return c.ServiceURL(api, id, "databases", databaseName)
}

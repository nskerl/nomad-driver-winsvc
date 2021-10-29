package winsvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	guid = "d42d7b18-691b-409a-94fd-4259a2b7e066"
	hash = "854d57551e79656159a0081054fbc08c6c648f86"
)

func TestGetService(t *testing.T) {
	assert := assert.New(t)

	client := ServiceClient{nil}

	var serviceName = "Nomad"
	if service, err := client.InspectService(serviceName); err != nil {
		t.Fatal("Error getting service: ", err)
	} else {
		assert.NotEmpty(service)
		assert.Equal(service.Name, "Nomad")
	}
}

//
///*func TestIsServiceRunning(t *testing.T) {
//	assert := assert.New(t)
//
//	var serviceName = "nomad.winsvc.hello-world.7ce1ade9-fee9-e966-94fd-4e40e22f3eec"
//
//	if isServiceRunning, err := IsServiceRunning(serviceName); err != nil {
//		t.Fatal("Failed to get isServiceRunning: ", err)
//	} else {
//		assert.True(isServiceRunning)
//	}
//}*/
//
//
//
////func TestGetServices(t *testing.T) {
////	assert := assert.New(t)
////
////
////	if result, err := getServices(); err != nil {
////		t.Fatal("Failed to getServices: ", err)
////	} else {
////		assert.GreaterOrEqual(len(result), 1)
////	}
////}
//
//func TestCreateService(t *testing.T) {
//	assert := assert.New(t)
//
//	serviceName := fmt.Sprintf("nomad.winsvc.%s.%s", "tests", guid)
//
//	serviceConfig := ServiceConfig{
//		Name:             serviceName,
//		Executable:   "c:\\temp\\echo-service.exe",
//		Args:             "",
//		DisplayName:      serviceName,
//		Description:      "Nomad managed service instance",
//		ServiceStartName: "",
//		Password:         "",
//	}
//
//	if err := createService(&serviceConfig); err !=nil {
//		t.Fatal("Failed to create service:", err)
//	}
//
//	// start it
//	if err := startService(serviceName); err != nil {
//		t.Fatal("Failed to start service:", err)
//	}
//	// inspect it
//	if service, err := getWmiService(serviceConfig.Name); err != nil {
//		t.Fatal("Error getting created service: ", err)
//	} else {
//		assert.NotEmpty(service)
//		assert.Equal(service.Name, serviceConfig.Name)
//	}
//
//
//}
//
//func TestStartService(t *testing.T) {
//	assert := assert.New(t)
//
//	serviceName := "Nomad"
//
//	if isServiceRunning, err := isServiceRunning(serviceName); err != nil {
//		t.Fatal("Failed to get isServiceRunning: ", err)
//	} else {
//		assert.True(isServiceRunning)
//	}
//}
//
//func TestStopService(t *testing.T) {
//	assert := assert.New(t)
//
//	serviceName := fmt.Sprintf("nomad.winsvc.%s.%s", "tests", guid)
//
//	stopService(serviceName)
//
//	if isServiceRunning, err := isServiceRunning(serviceName); err != nil {
//		t.Fatal("Failed to get isServiceRunning: ", err)
//	} else {
//		assert.False(isServiceRunning)
//	}
//}
//

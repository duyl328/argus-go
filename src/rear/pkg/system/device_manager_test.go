package system

import (
	"encoding/json"
	"testing"
)

func TestDeviceManagerScan(t *testing.T) {
	t.Log("Testing DeviceManager scanning functionality...")

	dm := NewDeviceManager()
	err := dm.ScanDevices()
	if err != nil {
		t.Fatalf("ScanDevices failed: %v", err)
	}

	devices := dm.GetDevices()
	if len(devices) == 0 {
		t.Error("No devices found - this is unexpected")
		return
	}

	t.Logf("Found %d storage devices:", len(devices))

	for i, device := range devices {
		t.Logf("Device %d:", i+1)
		t.Logf("  ID: %s", device.ID)
		t.Logf("  Name: %s", device.Name)
		t.Logf("  Mount Point: %s", device.MountPoint)
		t.Logf("  Device Path: %s", device.DevicePath)
		t.Logf("  Type: %s", device.Type)
		t.Logf("  File System: %s", device.FileSystem)
		t.Logf("  Label: %s", device.Label)
		t.Logf("  Is Ready: %t", device.IsReady)
		t.Logf("  Is Removable: %t", device.IsRemovable)
		if device.TotalSpace > 0 {
			t.Logf("  Total Space: %d bytes", device.TotalSpace)
			t.Logf("  Free Space: %d bytes", device.FreeSpace)
			t.Logf("  Used Space: %d bytes", device.UsedSpace)
		}
		t.Logf("  Created At: %s", device.CreatedAt.Format("2006-01-02 15:04:05"))
		t.Log("  ---")
	}
}

func TestDeviceManagerOperations(t *testing.T) {
	t.Log("Testing DeviceManager operations...")

	dm := NewDeviceManager()
	err := dm.ScanDevices()
	if err != nil {
		t.Fatalf("ScanDevices failed: %v", err)
	}

	devices := dm.GetDevices()
	if len(devices) == 0 {
		t.Skip("No devices found, skipping operations test")
		return
	}

	// Test GetDevice
	firstDevice := devices[0]
	foundDevice := dm.GetDevice(firstDevice.MountPoint)
	if foundDevice == nil {
		t.Errorf("GetDevice failed to find device: %s", firstDevice.MountPoint)
	} else {
		t.Logf("GetDevice found: %s -> %s", firstDevice.MountPoint, foundDevice.Name)
	}

	// Test GetDeviceByID
	foundByID := dm.GetDeviceByID(firstDevice.ID)
	if foundByID == nil {
		t.Errorf("GetDeviceByID failed to find device: %s", firstDevice.ID)
	} else {
		t.Logf("GetDeviceByID found: %s -> %s", firstDevice.ID, foundByID.Name)
	}

	// Test GetReadyDevices
	readyDevices := dm.GetReadyDevices()
	t.Logf("Found %d ready devices", len(readyDevices))

	// Test GetDevicesByType
	for _, deviceType := range []DeviceType{DeviceTypeFixed, DeviceTypeRemovable, DeviceTypeNetwork} {
		typeDevices := dm.GetDevicesByType(deviceType)
		if len(typeDevices) > 0 {
			t.Logf("Found %d devices of type %s", len(typeDevices), deviceType)
		}
	}

	// Test GetLargestFreeSpaceDevice
	largestDevice := dm.GetLargestFreeSpaceDevice()
	if largestDevice != nil {
		t.Logf("Largest free space device: %s (%d bytes free)", largestDevice.Name, largestDevice.FreeSpace)
	}

	// Test IsMountPoint
	isMountPoint := dm.IsMountPoint(firstDevice.MountPoint)
	if !isMountPoint {
		t.Errorf("IsMountPoint should return true for %s", firstDevice.MountPoint)
	}

	// Test RefreshDevice
	err = dm.RefreshDevice(firstDevice.MountPoint)
	if err != nil {
		t.Logf("RefreshDevice warning: %v", err)
	} else {
		t.Logf("RefreshDevice successful for %s", firstDevice.MountPoint)
	}
}

func TestDeviceManagerJSON(t *testing.T) {
	t.Log("Testing DeviceManager JSON serialization...")

	dm := NewDeviceManager()
	err := dm.ScanDevices()
	if err != nil {
		t.Fatalf("ScanDevices failed: %v", err)
	}

	jsonData, err := dm.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	t.Logf("JSON representation:\n%s", string(jsonData))

	// Test deserialization
	var devices []StorageDevice
	err = json.Unmarshal(jsonData, &devices)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	t.Logf("Successfully deserialized %d devices", len(devices))
}

func TestGlobalDeviceManager(t *testing.T) {
	t.Log("Testing global DeviceManager functions...")

	err := ScanStorageDevices()
	if err != nil {
		t.Fatalf("ScanStorageDevices failed: %v", err)
	}

	devices := GetStorageDevices()
	t.Logf("Global manager found %d devices", len(devices))

	if len(devices) > 0 {
		firstDevice := devices[0]

		foundDevice := GetStorageDevice(firstDevice.MountPoint)
		if foundDevice == nil {
			t.Errorf("GetStorageDevice failed to find device: %s", firstDevice.MountPoint)
		}

		readyDevices := GetReadyStorageDevices()
		t.Logf("Found %d ready devices via global function", len(readyDevices))

		isMountPoint := IsStorageMountPoint(firstDevice.MountPoint)
		if !isMountPoint {
			t.Errorf("IsStorageMountPoint should return true for %s", firstDevice.MountPoint)
		}
	}
}
package rollout

func CloneArtifact(value Artifact) Artifact {
	clone := value
	clone.DeviceClasses = append([]string(nil), value.DeviceClasses...)
	clone.Labels = make(map[string]string, len(value.Labels))
	for key, item := range value.Labels {
		clone.Labels[key] = item
	}
	return clone
}

func CloneCampaign(value Campaign) Campaign {
	clone := value
	clone.DeviceIDs = append([]string(nil), value.DeviceIDs...)
	return clone
}

func CloneDevices(values []Device) []Device {
	return append([]Device(nil), values...)
}

func RestoreLabels(snapshot map[string]string) map[string]string {
	labels := make(map[string]string, len(snapshot))
	for key, value := range snapshot {
		labels[key] = value
	}
	return labels
}

package evaluation

import "github.com/amplitude/experiment-go-server/pkg/experiment"

func UserToContext(user *experiment.User) map[string]interface{} {
	if user == nil {
		return nil
	}
	context := make(map[string]interface{})
	userMap := make(map[string]interface{})
	if len(user.UserId) != 0 {
		userMap["user_id"] = user.UserId
	}
	if len(user.DeviceId) != 0 {
		userMap["device_id"] = user.DeviceId
	}
	if len(user.Country) != 0 {
		userMap["country"] = user.Country
	}
	if len(user.Region) != 0 {
		userMap["region"] = user.Region
	}
	if len(user.Dma) != 0 {
		userMap["dma"] = user.Dma
	}
	if len(user.City) != 0 {
		userMap["city"] = user.City
	}
	if len(user.Language) != 0 {
		userMap["language"] = user.Language
	}
	if len(user.Platform) != 0 {
		userMap["platform"] = user.Platform
	}
	if len(user.Version) != 0 {
		userMap["version"] = user.Version
	}
	if len(user.Os) != 0 {
		userMap["os"] = user.Os
	}
	if len(user.DeviceManufacturer) != 0 {
		userMap["device_manufacturer"] = user.DeviceManufacturer
	}
	if len(user.DeviceBrand) != 0 {
		userMap["device_brand"] = user.DeviceBrand
	}
	if len(user.DeviceModel) != 0 {
		userMap["device_model"] = user.DeviceModel
	}
	if len(user.Carrier) != 0 {
		userMap["carrier"] = user.Carrier
	}
	if len(user.Library) != 0 {
		userMap["library"] = user.Library
	}
	if len(user.UserProperties) != 0 {
		userMap["user_properties"] = user.UserProperties
	}
	context["user"] = userMap
	return context
}

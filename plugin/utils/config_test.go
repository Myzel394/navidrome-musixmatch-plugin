package utils

import (
	"testing"

	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func TestMobileIdentityConfigDefaults(t *testing.T) {
	pdk.PDKMock.On("GetConfig", ConfigKeyMobileUserAgent).Return("", false).Once()
	pdk.PDKMock.On("GetConfig", ConfigKeyMobileAppVersion).Return("", false).Once()
	pdk.PDKMock.On("GetConfig", ConfigKeyMobileAppID).Return("", false).Once()
	t.Cleanup(func() { pdk.PDKMock.ExpectedCalls = nil; pdk.PDKMock.Calls = nil })

	if got := ConfigMobileUserAgent(); got != DefaultMobileUserAgent {
		t.Fatalf("ConfigMobileUserAgent()=%q want %q", got, DefaultMobileUserAgent)
	}
	if got := ConfigMobileAppVersion(); got != DefaultMobileAppVersion {
		t.Fatalf("ConfigMobileAppVersion()=%q want %q", got, DefaultMobileAppVersion)
	}
	if got := ConfigMobileAppID(); got != DefaultMobileAppID {
		t.Fatalf("ConfigMobileAppID()=%q want %q", got, DefaultMobileAppID)
	}
}

func TestMobileIdentityConfigOverrides(t *testing.T) {
	pdk.PDKMock.On("GetConfig", ConfigKeyMobileUserAgent).Return("CustomUA", true).Once()
	pdk.PDKMock.On("GetConfig", ConfigKeyMobileAppVersion).Return("1.2.3", true).Once()
	pdk.PDKMock.On("GetConfig", ConfigKeyMobileAppID).Return("custom-app", true).Once()
	t.Cleanup(func() { pdk.PDKMock.ExpectedCalls = nil; pdk.PDKMock.Calls = nil })

	if got := ConfigMobileUserAgent(); got != "CustomUA" {
		t.Fatalf("ConfigMobileUserAgent()=%q", got)
	}
	if got := ConfigMobileAppVersion(); got != "1.2.3" {
		t.Fatalf("ConfigMobileAppVersion()=%q", got)
	}
	if got := ConfigMobileAppID(); got != "custom-app" {
		t.Fatalf("ConfigMobileAppID()=%q", got)
	}
}

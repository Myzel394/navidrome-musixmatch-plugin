package musixmatch

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
)

const (
	mobileGUIDVariantRandom = "random_guid"
	mobileGUIDVariantNone   = "no_guid"
)

type mobileGUIDAssignment struct {
	Variant string
	GUID    string
}

var mobileExperimentRandomReader io.Reader = rand.Reader

func chooseMobileGUIDAssignment() mobileGUIDAssignment {
	var bit [1]byte
	if _, err := io.ReadFull(mobileExperimentRandomReader, bit[:]); err != nil {
		utils.LogErrorf("mobile API GUID experiment random selection failed: %v", err)
		return mobileGUIDAssignment{Variant: mobileGUIDVariantNone}
	}
	if bit[0]&1 == 0 {
		return mobileGUIDAssignment{Variant: mobileGUIDVariantNone}
	}
	guid, err := newMobileInstallGUID()
	if err != nil {
		utils.LogErrorf("mobile API GUID experiment GUID generation failed: %v", err)
		return mobileGUIDAssignment{Variant: mobileGUIDVariantNone}
	}
	return mobileGUIDAssignment{Variant: mobileGUIDVariantRandom, GUID: guid}
}

func newMobileInstallGUID() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(mobileExperimentRandomReader, b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

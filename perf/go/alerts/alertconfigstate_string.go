// Code generated by "stringer -type=AlertConfigState"; DO NOT EDIT

package alerts

import "fmt"

const _AlertConfigState_name = "ACTIVEDELETEDEOL"

var _AlertConfigState_index = [...]uint8{0, 6, 13, 16}

func (i ConfigState) String() string {
	if i < 0 || i >= ConfigState(len(_AlertConfigState_index)-1) {
		return fmt.Sprintf("AlertConfigState(%d)", i)
	}
	return _AlertConfigState_name[_AlertConfigState_index[i]:_AlertConfigState_index[i+1]]
}
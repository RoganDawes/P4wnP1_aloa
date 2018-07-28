package main

import (
	"github.com/mame82/hvue"
)

/*
type CompModalData struct {
	*js.Object
	ShowModal bool `js:"showModal"`
}

func NewCompModalData(vm *hvue.VM) interface{} {
	d:= &CompModalData{Object:O()}
	d.ShowModal = false
	return d
}
*/
func InitCompModal() {
	hvue.NewComponent(
		"modal",
//		hvue.DataFunc(NewCompModalData),
		hvue.Template(compModalTemplate),
	)
}

const compModalTemplate = `
 <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="modal-container">

          <div class="modal-header">
            <slot name="header">
              Modal window header
            </slot>
          </div>

          <div class="modal-body">
            <slot name="body">
              body
            </slot>
          </div>

          <div class="modal-footer">
            <slot name="footer">
				footer
              <button class="modal-default-button" @click="$emit('close')">
                OK
              </button>
            </slot>
          </div>
        </div>
      </div>
    </div>
  </transition>
`
package handlers

import (
	"k8s.io/klog/v2"
)

// InformerHandler is
type InformerHandler struct {
}

// NewInformerHandler is
func NewInformerHandler() *InformerHandler {
	return &InformerHandler{}
}

// OnAdd is
func (h *InformerHandler) OnAdd(obj interface{}) {
	klog.V(4).Infof("Added object %s", obj.GetName())
}

// OnUpdate is
func (h *InformerHandler) OnUpdate(old, new interface{}) {
	klog.V(4).Infof("Updated object %s", obj.GetName())
}

// OnDelete is
func (h *InformerHandler) OnDelete(obj interface{}) {
	klog.V(4).Infof("Deleted object %s", obj.GetName())
}

// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1

import (
	status "github.com/operator-framework/operator-sdk/pkg/status"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EditorStatus) DeepCopyInto(out *EditorStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EditorStatus.
func (in *EditorStatus) DeepCopy() *EditorStatus {
	if in == nil {
		return nil
	}
	out := new(EditorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupDB) DeepCopyInto(out *TupDB) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupDB.
func (in *TupDB) DeepCopy() *TupDB {
	if in == nil {
		return nil
	}
	out := new(TupDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TupDB) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupDBFrom) DeepCopyInto(out *TupDBFrom) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupDBFrom.
func (in *TupDBFrom) DeepCopy() *TupDBFrom {
	if in == nil {
		return nil
	}
	out := new(TupDBFrom)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupDBList) DeepCopyInto(out *TupDBList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TupDB, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupDBList.
func (in *TupDBList) DeepCopy() *TupDBList {
	if in == nil {
		return nil
	}
	out := new(TupDBList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TupDBList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupDBSpec) DeepCopyInto(out *TupDBSpec) {
	*out = *in
	out.From = in.From
	out.To = in.To
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupDBSpec.
func (in *TupDBSpec) DeepCopy() *TupDBSpec {
	if in == nil {
		return nil
	}
	out := new(TupDBSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupDBStatus) DeepCopyInto(out *TupDBStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]status.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupDBStatus.
func (in *TupDBStatus) DeepCopy() *TupDBStatus {
	if in == nil {
		return nil
	}
	out := new(TupDBStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupDBTo) DeepCopyInto(out *TupDBTo) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupDBTo.
func (in *TupDBTo) DeepCopy() *TupDBTo {
	if in == nil {
		return nil
	}
	out := new(TupDBTo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupWAS) DeepCopyInto(out *TupWAS) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupWAS.
func (in *TupWAS) DeepCopy() *TupWAS {
	if in == nil {
		return nil
	}
	out := new(TupWAS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TupWAS) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupWASList) DeepCopyInto(out *TupWASList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TupWAS, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupWASList.
func (in *TupWASList) DeepCopy() *TupWASList {
	if in == nil {
		return nil
	}
	out := new(TupWASList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TupWASList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupWASSpec) DeepCopyInto(out *TupWASSpec) {
	*out = *in
	out.From = in.From
	out.To = in.To
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupWASSpec.
func (in *TupWASSpec) DeepCopy() *TupWASSpec {
	if in == nil {
		return nil
	}
	out := new(TupWASSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupWASStatus) DeepCopyInto(out *TupWASStatus) {
	*out = *in
	if in.LastAnalyzeStartTime != nil {
		in, out := &in.LastAnalyzeStartTime, &out.LastAnalyzeStartTime
		*out = (*in).DeepCopy()
	}
	if in.LastAnalyzeCompletionTime != nil {
		in, out := &in.LastAnalyzeCompletionTime, &out.LastAnalyzeCompletionTime
		*out = (*in).DeepCopy()
	}
	if in.LastBuildStartTime != nil {
		in, out := &in.LastBuildStartTime, &out.LastBuildStartTime
		*out = (*in).DeepCopy()
	}
	if in.LastBuildCompletionTime != nil {
		in, out := &in.LastBuildCompletionTime, &out.LastBuildCompletionTime
		*out = (*in).DeepCopy()
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]status.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Editor != nil {
		in, out := &in.Editor, &out.Editor
		*out = new(EditorStatus)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupWASStatus.
func (in *TupWASStatus) DeepCopy() *TupWASStatus {
	if in == nil {
		return nil
	}
	out := new(TupWASStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupWasFrom) DeepCopyInto(out *TupWasFrom) {
	*out = *in
	out.Git = in.Git
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupWasFrom.
func (in *TupWasFrom) DeepCopy() *TupWasFrom {
	if in == nil {
		return nil
	}
	out := new(TupWasFrom)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupWasGit) DeepCopyInto(out *TupWasGit) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupWasGit.
func (in *TupWasGit) DeepCopy() *TupWasGit {
	if in == nil {
		return nil
	}
	out := new(TupWasGit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupWasImage) DeepCopyInto(out *TupWasImage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupWasImage.
func (in *TupWasImage) DeepCopy() *TupWasImage {
	if in == nil {
		return nil
	}
	out := new(TupWasImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TupWasTo) DeepCopyInto(out *TupWasTo) {
	*out = *in
	out.Image = in.Image
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TupWasTo.
func (in *TupWasTo) DeepCopy() *TupWasTo {
	if in == nil {
		return nil
	}
	out := new(TupWasTo)
	in.DeepCopyInto(out)
	return out
}

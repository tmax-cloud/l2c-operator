// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodeIssue) DeepCopyInto(out *CodeIssue) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodeIssue.
func (in *CodeIssue) DeepCopy() *CodeIssue {
	if in == nil {
		return nil
	}
	out := new(CodeIssue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2c) DeepCopyInto(out *L2c) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2c.
func (in *L2c) DeepCopy() *L2c {
	if in == nil {
		return nil
	}
	out := new(L2c)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *L2c) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cDb) DeepCopyInto(out *L2cDb) {
	*out = *in
	if in.From != nil {
		in, out := &in.From, &out.From
		*out = new(L2cDbFrom)
		**out = **in
	}
	if in.To != nil {
		in, out := &in.To, &out.To
		*out = new(L2cDbTo)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cDb.
func (in *L2cDb) DeepCopy() *L2cDb {
	if in == nil {
		return nil
	}
	out := new(L2cDb)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cDbFrom) DeepCopyInto(out *L2cDbFrom) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cDbFrom.
func (in *L2cDbFrom) DeepCopy() *L2cDbFrom {
	if in == nil {
		return nil
	}
	out := new(L2cDbFrom)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cDbTo) DeepCopyInto(out *L2cDbTo) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cDbTo.
func (in *L2cDbTo) DeepCopy() *L2cDbTo {
	if in == nil {
		return nil
	}
	out := new(L2cDbTo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cGit) DeepCopyInto(out *L2cGit) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cGit.
func (in *L2cGit) DeepCopy() *L2cGit {
	if in == nil {
		return nil
	}
	out := new(L2cGit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cImage) DeepCopyInto(out *L2cImage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cImage.
func (in *L2cImage) DeepCopy() *L2cImage {
	if in == nil {
		return nil
	}
	out := new(L2cImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cList) DeepCopyInto(out *L2cList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]L2c, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cList.
func (in *L2cList) DeepCopy() *L2cList {
	if in == nil {
		return nil
	}
	out := new(L2cList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *L2cList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cSpec) DeepCopyInto(out *L2cSpec) {
	*out = *in
	if in.Was != nil {
		in, out := &in.Was, &out.Was
		*out = new(L2cWas)
		(*in).DeepCopyInto(*out)
	}
	if in.Db != nil {
		in, out := &in.Db, &out.Db
		*out = new(L2cDb)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cSpec.
func (in *L2cSpec) DeepCopy() *L2cSpec {
	if in == nil {
		return nil
	}
	out := new(L2cSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cStatus) DeepCopyInto(out *L2cStatus) {
	*out = *in
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.TableRowCondition, len(*in))
		copy(*out, *in)
	}
	if in.TaskStatus != nil {
		in, out := &in.TaskStatus, &out.TaskStatus
		*out = make([]L2cTaskStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SonarIssues != nil {
		in, out := &in.SonarIssues, &out.SonarIssues
		*out = make([]CodeIssue, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cStatus.
func (in *L2cStatus) DeepCopy() *L2cStatus {
	if in == nil {
		return nil
	}
	out := new(L2cStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cTaskStatus) DeepCopyInto(out *L2cTaskStatus) {
	*out = *in
	in.TaskRunStatus.DeepCopyInto(&out.TaskRunStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cTaskStatus.
func (in *L2cTaskStatus) DeepCopy() *L2cTaskStatus {
	if in == nil {
		return nil
	}
	out := new(L2cTaskStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cWas) DeepCopyInto(out *L2cWas) {
	*out = *in
	if in.From != nil {
		in, out := &in.From, &out.From
		*out = new(L2cWasFrom)
		(*in).DeepCopyInto(*out)
	}
	if in.To != nil {
		in, out := &in.To, &out.To
		*out = new(L2cWasTo)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cWas.
func (in *L2cWas) DeepCopy() *L2cWas {
	if in == nil {
		return nil
	}
	out := new(L2cWas)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cWasFrom) DeepCopyInto(out *L2cWasFrom) {
	*out = *in
	if in.Git != nil {
		in, out := &in.Git, &out.Git
		*out = new(L2cGit)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cWasFrom.
func (in *L2cWasFrom) DeepCopy() *L2cWasFrom {
	if in == nil {
		return nil
	}
	out := new(L2cWasFrom)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L2cWasTo) DeepCopyInto(out *L2cWasTo) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(L2cImage)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L2cWasTo.
func (in *L2cWasTo) DeepCopy() *L2cWasTo {
	if in == nil {
		return nil
	}
	out := new(L2cWasTo)
	in.DeepCopyInto(out)
	return out
}

package filter

func (in *Cdump) DeepCopyInto(out *Cdump) {
	*out = *in
	if in.URLs != nil {
		out.URLs = make([]string, 0, len(in.URLs))
		for _, i := range in.URLs {
			out.URLs = append(out.URLs, i)
		}
	}
}

func (in *Creplacement) DeepCopyInto(out *Creplacement) {
	*out = *in
	if in.Urls != nil {
		out.Urls = make([]string, 0, len(in.Urls))
		for _, i := range in.Urls {
			out.Urls = append(out.Urls, i)
		}
	}
}
func (in *Cheader) DeepCopyInto(out *Cheader) {
	*out = *in
}

func (in *Caction) DeepCopyInto(out *Caction) {
	*out = *in
	if in.Header != nil {
		in, out := &in.Header, &out.Header
		*out = make([]Cheader, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Replace != nil {
		in, out := &in.Replace, &out.Replace
		*out = make([]Creplacement, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

package icons

import "fyne.io/fyne/v2"

func CheckIcon() fyne.Resource {
	return fyne.NewStaticResource("check", Check)
}

func ErrorIcon() fyne.Resource {
	return fyne.NewStaticResource("error", Error)
}

func RunningIcon() fyne.Resource {
	return fyne.NewStaticResource("running", Running)
}

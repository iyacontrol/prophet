package main

const (
	K8sDeployStageInitialization = iota
	K8sDeployStageCanary
	K8sDeployStageRollBack
	K8sDeployStageRollup
)

const (
	CronJobApiVersion = "batch/v1beta1"
	CronJobkind       = "CronJob"
)

const (
	ScheduleCronExp = "0 d% * * *"
)

const  ImagePullSecret  = "default-secret"

const (
	EnvHpa = "CRONHPA_Name"
	EnvCapacity = "CRONHPA_CAPACITY"

	EnvProphetImage = "PROPHET_IMAGE"
	EnvProphetAccount = "PROPHET_ACCOUNT"
)
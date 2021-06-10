package config

const (

	// SparkRoleLabel is the driver/executor label set by the operator/spark-distribution on the driver/executors Pods.
	SparkRoleLabel = "spark-role"

	// LabelAnnotationPrefix is the prefix of every labels and annotations added by the controller.
	LabelAnnotationPrefix = "sparkoperator.k8s.io/"

	// LaunchedBySparkOperatorLabel is a label on Spark pods launched through the Spark Operator.
	LaunchedBySparkOperatorLabel = LabelAnnotationPrefix + "launched-by-spark-operator"
)

package logger

func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

func Error(val string) Field {
	return Field{Key: "error", Value: val}
}

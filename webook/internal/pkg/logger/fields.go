package logger

func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

func Error(val string) Field {
	return Field{Key: "error", Value: val}
}

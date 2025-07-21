package logs

//func GrpcLoggerUnaryInterceptor() grpc.UnaryServerInterceptor {
//	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
//		startTime := time.Now()
//		// ... 获取客户端信息 ...
//
//		resp, err := handler(ctx, req) // 处理请求
//
//		endTime := time.Now()
//		latency := int(endTime.Sub(startTime).Milliseconds())
//
//		// 构造日志条目
//		logEntry := LogEntry{
//			Service:   "grpc",
//			Method:    info.FullMethod,
//			ClientID:  clientID,
//			Latency:   latency,
//			Timestamp: endTime,
//			Error:     err.Error(),
//		}
//
//		LogAsync(logEntry) // 异步记录日志
//		return resp, err
//	}
//}
//
//func GinLoggerMiddleware() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		startTime := time.Now()
//		c.Next() // 处理请求
//		endTime := time.Now()
//
//		logEntry := LogEntry{
//			Service:    "gin",
//			Method:     c.Request.URL.Path,
//			StatusCode: c.Writer.Status(),
//			Latency:    int(endTime.Sub(startTime).Milliseconds()),
//			Timestamp:  endTime,
//		}
//
//		if len(c.Errors) > 0 {
//			logEntry.Error = c.Errors.String()
//		}
//
//		LogAsync(logEntry) // 异步记录
//	}
//}
//
//var logChannel = make(chan LogEntry, 100) // 缓冲通道
//// 异步记录入口
//func LogAsync(log LogEntry) {
//	logChannel <- log // 非阻塞提交
//}
//
//// 日志处理器协程
//func logProcessor() {
//	var logs []LogEntry
//	ticker := time.NewTicker(5 * time.Second)
//
//	for {
//		select {
//		case logEntry := <-logChannel:
//			logs = append(logs, logEntry)
//			if len(logs) >= 10 { // 批量写入
//				insertDbLogs(logs)
//				logs = []LogEntry{}
//			}
//		case <-ticker.C: // 定时刷新
//			if len(logs) > 0 {
//				insertDbLogs(logs)
//				logs = []LogEntry{}
//			}
//		}
//	}
//}
//
//// 批量写入数据库
//func insertDbLogs(logs []LogEntry) {
//	query := "INSERT INTO Logs (...) VALUES "
//	for i, log := range logs {
//		if i > 0 {
//			query += ","
//		}
//		query += fmt.Sprintf("('%s', %d, ...)", log.Method, log.Latency)
//	}
//	db.Exec(query)
//}
//
//func CloseLogging() {
//	close(logChannel) // 关闭通道
//	wg.Wait()         // 等待处理完成
//}

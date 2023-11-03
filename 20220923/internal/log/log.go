package log

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

const logTimeFmt = "2006-01-02 15:04:05.000"

// 指定配置文件位置。应该写在各启动服务中，为了简化写在这里
func init() {
	initConfig()
	initLog()
}

// 初始化配置文件
func initConfig() {
	viper.AutomaticEnv()
	viper.SetConfigName("conf")
	viper.SetConfigType("toml")
	viper.AddConfigPath("internal/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

// 替换全局日志配置
func initLog() {
	// 设置日志输出的格式
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString("[" + t.Format(logTimeFmt) + "]")
	}
	cfg.EncodeCaller = func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString("[" + caller.TrimmedPath() + "]")
	}
	cfg.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString("[" + level.CapitalString() + "]")
	}

	// 默认为 info 级别
	level := zapcore.InfoLevel
	// 设置日志级别
	if v := viper.GetString("log.level"); v != "" {
		if err := level.UnmarshalText([]byte(v)); err != nil {
			panic("log.level error")
		}
	}

	ws := make([]zapcore.WriteSyncer, 0)
	// 日志写入控制台
	ws = append(ws, zapcore.AddSync(os.Stdout))
	// 日志写入文件
	if viper.GetBool("log.file") {
		// 切割日志
		hook := &lumberjack.Logger{
			Filename:   "log.log", //日志文件路径
			MaxSize:    30,        // 每个日志文件保存的大小 单位:M
			MaxAge:     7,         // 文件最多保存多少天
			MaxBackups: 30,        // 日志文件最多保存多少个备份
			Compress:   false,     // 是否压缩
		}
		ws = append(ws, zapcore.AddSync(hook))
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zap.CombineWriteSyncers(ws...),
		zap.NewAtomicLevelAt(level),
	)
	defer core.Sync()

	// 全局替换
	zap.ReplaceGlobals(zap.New(core, zap.AddCaller()))
}

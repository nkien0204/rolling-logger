# rolling-logger
A helpful tool for separating log file by time!
## How to use
```bash
package main

import (
	"github.com/nkien0204/rolling-logger/rolling"
)

func main() {
	logger := rolling.New()
	defer logger.Sync()

	logger.Info("hello logger")
	logger.Error("got error")
	logger.Debug("this is debug")
}
```
Some basic configurations are on the `config.yaml` file. So make sure that it available to load this configuration.

**Let's take a look at `config.yaml`:**
- `log_rotation_time`: (`day`|`hour`|`min`) for "daily", "hourly" or "every minute" log file separation (default is `hour`).
- `log_info_dir`/`log_info_name`: location of log files which have the level **greater or equal to INFO**.
- `log_debug_dir`/`log_debug_name`: location of log files which have the level **less than INFO**.

**Log level order:** `DEBUG` < `INFO` < `WARN` < `ERROR` < `PANIC` < `FATAL`

**Tracking the latest log:** `logger.log` and `logger-debug.log` (in case using `DEBUG` log level)

## Dependencies
- [uber-go/zap](https://github.com/uber-go/zap)
- [lestrrat-go/strftime](https://github.com/lestrrat-go/strftime)
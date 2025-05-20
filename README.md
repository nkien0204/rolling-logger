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
- `output`: choose output type between `console` and `file`
- `log_rotation_time`: (`day`|`hour`|`min`) for "daily", "hourly" or "every minute" log file separation (default is `hour`).
- `log_dir`/`log_file_name`: location of log files which match the level from `log_level_min` to `log_level_max`.
- `log_level_min` <= log level <= `log_level_max`

**Log level order:** `debug` < `info` < `warn` < `error`

**Tracking the latest log:** `logger.log`

## Dependencies
- [uber-go/zap](https://github.com/uber-go/zap)
- [lestrrat-go/strftime](https://github.com/lestrrat-go/strftime)

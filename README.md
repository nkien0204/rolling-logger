# rolling-logger
A helpful tool for separating log file by time!
## How to use
```bash
package main

import (
	"fmt"

	"github.com/nkien0204/rolling-logger/rolling"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("load env error: ", err.Error())
		panic(err)
	}
	logger := rolling.New()
	defer logger.Sync()

	logger.Info("hello logger")
	logger.Error("got error")
	logger.Debug("this is debug")
}
```
Some basic configurations are on the `.env` file. So make sure that it available to load this environment.

Let's take a look at `.env`:
- `LOG_ROTATION_TIME`: (`day`|`hour`|`min`) for "daily", "hourly" or "every minute" log file separation (default is `hour`).
- `LOG_INFO_DIR`/`LOG_INFO_NAME`: location of log files which have the level are greater or equal to **INFO**.
- `LOG_DEBUG_DIR`/`LOG_DEBUG_NAME`: location of log files which have the level less than **INFO**.

Log level order: `DEBUG` < `INFO` < `WARN` < `ERROR` < `PANIC` < `FATAL`

## Dependencies
- [joho/godotenv](https://github.com/joho/godotenv)
- [uber-go/zap](https://github.com/uber-go/zap)
- [lestrrat-go/strftime](https://github.com/lestrrat-go/strftime)
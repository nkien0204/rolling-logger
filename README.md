# rolling-logger

## How to use
```bash
package main

import (
	"fmt"

	"github.com/nkien0204/rolling-logger/logger"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("load env error: ", err.Error())
		panic(err)
	}
	logger := logger.New()
	defer logger.Sync()

	logger.Info("hello logger")
	logger.Error("got error")
}
```
Some basic configurations are on the `.env` file. So make sure that it available to load this environment.

Let take a look at `.env`:
- `LOG_ROTATION_TIME`: (`day`|`hour`|`min`) for "daily", "hourly" or "every minute" log file separation (default is `hour`).
- `LOG_FILE`: your log file location (default is `log/logger.log`). 

## Dependencies
- [joho/godotenv](https://github.com/joho/godotenv)
- [uber-go/zap](https://github.com/uber-go/zap)
- [strftime](https://github.com/lestrrat-go/strftime)
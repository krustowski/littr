# Apache JMeter

##  installation (linux amd64)

+ go to [Download JMeter](https://jmeter.apache.org/download_jmeter.cgi), download binary in `.tgz`, open sha512 link there, and to verify checksums by running something like this:

```
sha512sum Downloads/apache-jmeter-5.6.3.tgz | grep 5978a1a35edb5a7d428e270564ff49d2b1b257a65e17a759d259a9283fc17093e522fe46f474a043864aea6910683486340706d745fcdf3db1505fd71e689083
```

+ unpack and install it somewhere (e.g. to your HOME directory, say, `~/jmeter`)

## usage

+ fill your credentials in `auth.json`
+ and run the infinite loop (CTRL+C to stop):

```
jmeter -n -t ./litter_load_plan.jmx
```

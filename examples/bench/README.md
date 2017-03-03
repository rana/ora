# Benchmark ora vs. oci8
Compile a simple SELECT benchmark with `github.com/mattn/go-oci8` and `gopkg.in/rana/ora.v4`, and compare them:

	./run.sh <user/passw@host:port/sid>

## Result

```
$ ./run.sh
Compile ...
Benchmarks on DSN=boma/boma@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=p520.unosoft.local)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=cigf.unosoft.local))) ...
go-oci8
2016/03/30 11:10:37 Write CPU profile to "oci8.pprof".
2016/03/30 11:10:42 Iterated 5000 rows in 3.518387308s.
2016/03/30 11:10:42        1    3518345504 ns/op           0.00 MB/s
Entering interactive mode (type "help" for commands)
(pprof) (pprof) 20ms of 20ms total (  100%)
      flat  flat%   sum%        cum   cum%
         0     0%     0%       20ms   100%  database/sql.(*Rows).Next
         0     0%     0%       20ms   100%  github.com/mattn/go-oci8.(*OCI8Rows).Next
         0     0%     0%       20ms   100%  github.com/mattn/go-oci8._Cfunc_OCIStmtFetch
         0     0%     0%       20ms   100%  main.BenchmarkIter
      20ms   100%   100%       20ms   100%  runtime.cgocall
         0     0%   100%       20ms   100%  runtime.goexit
         0     0%   100%       20ms   100%  testing.(*B).launch
         0     0%   100%       20ms   100%  testing.(*B).runN
(pprof) ora
2016/03/30 11:10:42 Write CPU profile to "ora.pprof".
2016/03/30 11:10:56 Iterated 5000 rows in 12.527146357s.
2016/03/30 11:10:56        1    12527036391 ns/op          0.00 MB/s
Entering interactive mode (type "help" for commands)
(pprof) (pprof) 110ms of 110ms total (  100%)
      flat  flat%   sum%        cum   cum%
         0     0%     0%      110ms   100%  database/sql.(*Rows).Next
         0     0%     0%      110ms   100%  gopkg.in/rana/ora%2ev3.(*DrvQueryResult).Next
         0     0%     0%      110ms   100%  gopkg.in/rana/ora%2ev3.(*Rset).beginRow
         0     0%     0%      110ms   100%  gopkg.in/rana/ora%2ev3._Cfunc_OCIStmtFetch2
         0     0%     0%      110ms   100%  main.BenchmarkIter
     110ms   100%   100%      110ms   100%  runtime.cgocall
         0     0%   100%      110ms   100%  runtime.goexit
         0     0%   100%      110ms   100%  testing.(*B).launch
         0     0%   100%      110ms   100%  testing.(*B).runN
(pprof) :tgulacsi@tgulacsi-laptop: ~/src/gopkg.in/rana/ora.v4/examples/bench
```

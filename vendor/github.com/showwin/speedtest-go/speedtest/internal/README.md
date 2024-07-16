# Issue #192

<img width="1172" alt="SpeedTest-Go (1)" src="https://github.com/showwin/speedtest-go/assets/30739857/ecced0e9-830c-42d6-aa8b-e3dcf8d124d6">

1. Use welford alg to quickly calculate standard deviation and mean.
2. The welford alg integrated moving window feature, This allows us to ignore early data with excessive volatility.
3. Use the coefficient of variation(c.v) to reflect the confidence of the test result datasets.
4. When the data becomes stable(converge), the c.v value will become smaller. When the c.v < 0.05, we terminate this test. We set the tolerance condition as the window buffer being more than half filled and triggering more than five times with c.v < 0.05.
5. Perform EWMA operation on real-time global average, and use c.v as part of the EWMA feedback parameter.
6. The ewma value calculated is the result value of our test.
7. When the test data converge quickly, we can stop early and speed up the testing process. Of course this depends on network/device conditions.

doc:
	@echo ">>> http://localhost:6060/pkg/github.com/domahidizoltan/go-steps?m=all <<<"
	godoc -http=:6060 -play

run-benchmarks:
	@rm -f tmp/native.txt tmp/transformer.txt tmp/benchstat.txt
	go test -bench=BenchmarkNative -count=10 test/benchmarks_test.go > tmp/native.txt
	go test -bench=BenchmarkTransformer -count=10 test/benchmarks_test.go > tmp/transformer.txt
	@sed -i 's/Native//g' tmp/native.txt
	@sed -i 's/Transformer//g' tmp/transformer.txt
	@echo ""
	@benchstat tmp/native.txt tmp/transformer.txt > tmp/benchstat.txt
	@cat tmp/benchstat.txt

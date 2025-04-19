# 安装observe组件到k8s集群
observe:
	@$(MAKE) -C deploy/observe $(filter-out $@,$(MAKECMDGOALS))

# 生成grpc 代码
gen:
	@$(MAKE) -C api generate

# 避免 “No rule to make target 'grafana'” 等错误，需要一个空目标来吞掉子目标。
%:
    @:
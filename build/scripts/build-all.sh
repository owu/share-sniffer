#!/bin/bash

# 批处理执行所有bash脚本
current_dir=$(dirname "$0")

# 获取当前目录下所有.sh脚本，排除build-all.sh本身
scripts=$(find "$current_dir" -name "*.sh" -type f | grep -v "build-all.sh")

# 依次执行每个脚本
for script in $scripts; do
    echo "============================================="
    echo "正在执行: $(basename "$script")"
    echo "============================================="
    
    if bash "$script"; then
        echo "成功执行: $(basename "$script")"
    else
        echo "执行失败: $(basename "$script")"
        echo "错误代码: $?"
    fi
    
    echo ""
done

echo "============================================="
echo "所有脚本执行完毕"
echo "============================================="

import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import {
    GetConfig,
    SaveConfig,
    GetActiveWindow,
    GetCurrentInput,
    GetAvailableInputs,
    SwitchInput,
    AddRule,
    UpdateRule,
    DeleteRule,
    TestRule,
    IsRunning,
    SetRunning
} from '../wailsjs/go/main/App';

// 全局变量
let config = null;
let availableInputs = [];
let isRunning = false;

// 初始化应用
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

async function initializeApp() {
    document.getElementById('logo').src = logo;
    renderApp();

    // 加载数据
    await loadConfig();
    await loadAvailableInputs();
    await loadCurrentStatus();

    // 定期更新状态
    setInterval(updateStatus, 2000);
}

// 渲染主界面
function renderApp() {
    document.querySelector('#app').innerHTML = `
        <div class="app-container">
            <header class="app-header">
                <img id="logo" class="logo" alt="Logo">
                <h1>输入法自动切换</h1>
                <div class="status-toggle">
                    <label class="switch">
                        <input type="checkbox" id="runningToggle" onchange="toggleRunning()">
                        <span class="slider round"></span>
                    </label>
                    <span id="statusText">已启用</span>
                </div>
            </header>

            <main class="app-main">
                <!-- 当前状态 -->
                <section class="status-section">
                    <h2>当前状态</h2>
                    <div class="status-grid">
                        <div class="status-item">
                            <label>当前应用:</label>
                            <span id="currentApp">获取中...</span>
                        </div>
                        <div class="status-item">
                            <label>当前输入法:</label>
                            <span id="currentInput">获取中...</span>
                        </div>
                    </div>
                </section>

                <!-- 规则配置 -->
                <section class="rules-section">
                    <h2>切换规则</h2>
                    <div class="rules-controls">
                        <button class="btn btn-primary" onclick="showAddRuleDialog()">添加规则</button>
                        <button class="btn btn-secondary" onclick="refreshConfig()">刷新</button>
                    </div>
                    <div class="rules-list" id="rulesList">
                        <!-- 规则列表将在这里动态生成 -->
                    </div>
                </section>

                <!-- 通用设置 -->
                <section class="settings-section">
                    <h2>通用设置</h2>
                    <div class="settings-grid">
                        <div class="setting-item">
                            <label for="checkInterval">检查间隔 (毫秒):</label>
                            <input type="number" id="checkInterval" min="100" max="5000" step="100">
                        </div>
                        <div class="setting-item">
                            <label for="switchDelay">切换延迟 (毫秒):</label>
                            <input type="number" id="switchDelay" min="0" max="1000" step="50">
                        </div>
                        <div class="setting-item">
                            <label>
                                <input type="checkbox" id="showNotifications">
                                显示通知
                            </label>
                        </div>
                        <div class="setting-item">
                            <label>
                                <input type="checkbox" id="enableLogging">
                                启用日志
                            </label>
                        </div>
                    </div>
                    <div class="settings-actions">
                        <button class="btn btn-primary" onclick="saveSettings()">保存设置</button>
                    </div>
                </section>
            </main>
        </div>

        <!-- 添加规则对话框 -->
        <div class="modal" id="addRuleModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h3>添加切换规则</h3>
                    <span class="close" onclick="hideAddRuleDialog()">&times;</span>
                </div>
                <div class="modal-body">
                    <div class="form-group">
                        <label for="ruleAppName">应用名称:</label>
                        <input type="text" id="ruleAppName" placeholder="例如: com.apple.Safari">
                        <button class="btn btn-small" onclick="detectCurrentApp()">检测当前应用</button>
                    </div>
                    <div class="form-group">
                        <label for="ruleWindowName">窗口名称 (可选):</label>
                        <input type="text" id="ruleWindowName" placeholder="支持通配符 *">
                    </div>
                    <div class="form-group">
                        <label for="ruleInput">目标输入法:</label>
                        <select id="ruleInput">
                            <!-- 输入法选项将在这里动态生成 -->
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="rulePriority">优先级:</label>
                        <input type="number" id="rulePriority" min="1" max="100" value="1">
                        <small>数字越小优先级越高</small>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="ruleEnabled" checked>
                            启用规则
                        </label>
                    </div>
                </div>
                <div class="modal-footer">
                    <button class="btn btn-secondary" onclick="hideAddRuleDialog()">取消</button>
                    <button class="btn btn-primary" onclick="addRule()">添加</button>
                </div>
            </div>
        </div>
    `;
}

// 加载配置
async function loadConfig() {
    try {
        config = await GetConfig();
        renderRules();
        updateSettingsUI();
    } catch (err) {
        console.error('加载配置失败:', err);
        showMessage('加载配置失败: ' + err, 'error');
    }
}

// 加载可用输入法
async function loadAvailableInputs() {
    try {
        availableInputs = await GetAvailableInputs();
        updateInputSelects();
    } catch (err) {
        console.error('加载输入法列表失败:', err);
        showMessage('加载输入法列表失败: ' + err, 'error');
    }
}

// 加载当前状态
async function loadCurrentStatus() {
    try {
        isRunning = await IsRunning();
        document.getElementById('runningToggle').checked = isRunning;
        updateStatusText();
        await updateStatus();
    } catch (err) {
        console.error('加载状态失败:', err);
    }
}

// 更新状态
async function updateStatus() {
    try {
        const [window, input] = await Promise.all([
            GetActiveWindow(),
            GetCurrentInput()
        ]);

        document.getElementById('currentApp').textContent =
            window ? `${window.AppName} (${window.WindowName})` : '获取失败';
        document.getElementById('currentInput').textContent =
            input ? input.Name : '获取失败';
    } catch (err) {
        console.error('更新状态失败:', err);
    }
}

// 渲染规则列表
function renderRules() {
    if (!config || !config.rules) return;

    const rulesList = document.getElementById('rulesList');
    rulesList.innerHTML = '';

    config.rules.forEach((rule, index) => {
        const ruleElement = document.createElement('div');
        ruleElement.className = 'rule-item';
        ruleElement.innerHTML = `
            <div class="rule-info">
                <div class="rule-app">
                    <strong>应用:</strong> ${rule.app}
                    ${rule.window ? `<br><strong>窗口:</strong> ${rule.window}` : ''}
                </div>
                <div class="rule-input">
                    <strong>输入法:</strong> ${getInputName(rule.input)}
                </div>
                <div class="rule-priority">
                    <strong>优先级:</strong> ${rule.priority}
                </div>
                <div class="rule-enabled">
                    <label>
                        <input type="checkbox" ${rule.enabled ? 'checked' : ''}
                               onchange="toggleRule(${index}, this.checked)">
                        启用
                    </label>
                </div>
            </div>
            <div class="rule-actions">
                <button class="btn btn-small" onclick="testRule(${index})">测试</button>
                <button class="btn btn-small btn-edit" onclick="editRule(${index})">编辑</button>
                <button class="btn btn-small btn-delete" onclick="deleteRule(${index})">删除</button>
            </div>
        `;
        rulesList.appendChild(ruleElement);
    });
}

// 获取输入法名称
function getInputName(inputId) {
    const input = availableInputs.find(i => i.id === inputId);
    return input ? input.name : inputId;
}

// 更新输入法选择框
function updateInputSelects() {
    const selects = document.querySelectorAll('#ruleInput');
    selects.forEach(select => {
        select.innerHTML = availableInputs.map(input =>
            `<option value="${input.id}">${input.name}</option>`
        ).join('');
    });
}

// 更新设置界面
function updateSettingsUI() {
    if (!config || !config.general) return;

    document.getElementById('checkInterval').value = config.general.checkInterval;
    document.getElementById('switchDelay').value = config.general.switchDelay;
    document.getElementById('showNotifications').checked = config.general.showNotifications;
    document.getElementById('enableLogging').checked = config.general.enableLogging;
}

// 切换运行状态
async function toggleRunning() {
    const running = document.getElementById('runningToggle').checked;
    try {
        await SetRunning(running);
        isRunning = running;
        updateStatusText();
        showMessage(running ? '输入法自动切换已启用' : '输入法自动切换已禁用', 'success');
    } catch (err) {
        console.error('切换状态失败:', err);
        showMessage('切换状态失败: ' + err, 'error');
        // 恢复开关状态
        document.getElementById('runningToggle').checked = !running;
    }
}

// 更新状态文本
function updateStatusText() {
    document.getElementById('statusText').textContent = isRunning ? '已启用' : '已禁用';
}

// 显示添加规则对话框
function showAddRuleDialog() {
    document.getElementById('addRuleModal').style.display = 'block';
}

// 隐藏添加规则对话框
function hideAddRuleDialog() {
    document.getElementById('addRuleModal').style.display = 'none';
    clearRuleForm();
}

// 清空规则表单
function clearRuleForm() {
    document.getElementById('ruleAppName').value = '';
    document.getElementById('ruleWindowName').value = '';
    document.getElementById('rulePriority').value = '1';
    document.getElementById('ruleEnabled').checked = true;
}

// 检测当前应用
async function detectCurrentApp() {
    try {
        const window = await GetActiveWindow();
        if (window) {
            document.getElementById('ruleAppName').value = window.AppName;
            if (window.WindowName) {
                document.getElementById('ruleWindowName').value = window.WindowName;
            }
        }
    } catch (err) {
        console.error('检测当前应用失败:', err);
        showMessage('检测当前应用失败: ' + err, 'error');
    }
}

// 添加规则
async function addRule() {
    const rule = {
        app: document.getElementById('ruleAppName').value.trim(),
        window: document.getElementById('ruleWindowName').value.trim(),
        input: document.getElementById('ruleInput').value,
        priority: parseInt(document.getElementById('rulePriority').value),
        enabled: document.getElementById('ruleEnabled').checked
    };

    if (!rule.app) {
        showMessage('请输入应用名称', 'error');
        return;
    }

    if (!rule.input) {
        showMessage('请选择目标输入法', 'error');
        return;
    }

    try {
        await AddRule(rule);
        await loadConfig();
        hideAddRuleDialog();
        showMessage('规则添加成功', 'success');
    } catch (err) {
        console.error('添加规则失败:', err);
        showMessage('添加规则失败: ' + err, 'error');
    }
}

// 切换规则启用状态
async function toggleRule(index, enabled) {
    if (!config || !config.rules[index]) return;

    const rule = {...config.rules[index], enabled};
    try {
        await UpdateRule(index, rule);
        await loadConfig();
        showMessage('规则状态已更新', 'success');
    } catch (err) {
        console.error('更新规则失败:', err);
        showMessage('更新规则失败: ' + err, 'error');
    }
}

// 测试规则
async function testRule(index) {
    if (!config || !config.rules[index]) return;

    try {
        const [matched, window] = await TestRule(config.rules[index]);
        if (matched) {
            showMessage(`规则匹配成功! 当前应用: ${window.AppName}`, 'success');
        } else {
            showMessage(`规则不匹配当前应用: ${window.AppName}`, 'info');
        }
    } catch (err) {
        console.error('测试规则失败:', err);
        showMessage('测试规则失败: ' + err, 'error');
    }
}

// 编辑规则
function editRule(index) {
    if (!config || !config.rules[index]) return;

    const rule = config.rules[index];
    document.getElementById('ruleAppName').value = rule.app;
    document.getElementById('ruleWindowName').value = rule.window || '';
    document.getElementById('ruleInput').value = rule.input;
    document.getElementById('rulePriority').value = rule.priority;
    document.getElementById('ruleEnabled').checked = rule.enabled;

    // 修改对话框为编辑模式
    const modal = document.getElementById('addRuleModal');
    modal.querySelector('.modal-header h3').textContent = '编辑切换规则';
    modal.querySelector('.modal-footer .btn-primary').onclick = () => updateRule(index);
    modal.querySelector('.modal-footer .btn-primary').textContent = '更新';

    modal.style.display = 'block';
}

// 更新规则
async function updateRule(index) {
    const rule = {
        app: document.getElementById('ruleAppName').value.trim(),
        window: document.getElementById('ruleWindowName').value.trim(),
        input: document.getElementById('ruleInput').value,
        priority: parseInt(document.getElementById('rulePriority').value),
        enabled: document.getElementById('ruleEnabled').checked
    };

    if (!rule.app) {
        showMessage('请输入应用名称', 'error');
        return;
    }

    if (!rule.input) {
        showMessage('请选择目标输入法', 'error');
        return;
    }

    try {
        await UpdateRule(index, rule);
        await loadConfig();
        hideAddRuleDialog();
        resetAddRuleDialog();
        showMessage('规则更新成功', 'success');
    } catch (err) {
        console.error('更新规则失败:', err);
        showMessage('更新规则失败: ' + err, 'error');
    }
}

// 重置添加规则对话框
function resetAddRuleDialog() {
    const modal = document.getElementById('addRuleModal');
    modal.querySelector('.modal-header h3').textContent = '添加切换规则';
    modal.querySelector('.modal-footer .btn-primary').onclick = addRule;
    modal.querySelector('.modal-footer .btn-primary').textContent = '添加';
}

// 删除规则
async function deleteRule(index) {
    if (!config || !config.rules[index]) return;

    if (!confirm('确定要删除这条规则吗？')) return;

    try {
        await DeleteRule(index);
        await loadConfig();
        showMessage('规则删除成功', 'success');
    } catch (err) {
        console.error('删除规则失败:', err);
        showMessage('删除规则失败: ' + err, 'error');
    }
}

// 保存设置
async function saveSettings() {
    if (!config || !config.general) return;

    config.general.checkInterval = parseInt(document.getElementById('checkInterval').value);
    config.general.switchDelay = parseInt(document.getElementById('switchDelay').value);
    config.general.showNotifications = document.getElementById('showNotifications').checked;
    config.general.enableLogging = document.getElementById('enableLogging').checked;

    try {
        await SaveConfig(config);
        showMessage('设置保存成功', 'success');
    } catch (err) {
        console.error('保存设置失败:', err);
        showMessage('保存设置失败: ' + err, 'error');
    }
}

// 刷新配置
async function refreshConfig() {
    await loadConfig();
    showMessage('配置已刷新', 'success');
}

// 显示消息
function showMessage(message, type = 'info') {
    // 创建消息元素
    const messageElement = document.createElement('div');
    messageElement.className = `message message-${type}`;
    messageElement.textContent = message;

    // 添加到页面
    document.body.appendChild(messageElement);

    // 3秒后自动消失
    setTimeout(() => {
        if (messageElement.parentNode) {
            messageElement.parentNode.removeChild(messageElement);
        }
    }, 3000);
}

// 点击模态框外部关闭
window.onclick = function(event) {
    const modal = document.getElementById('addRuleModal');
    if (event.target === modal) {
        hideAddRuleDialog();
    }
}

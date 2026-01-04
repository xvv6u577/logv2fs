// ==================== 全局配置 ====================
const API_BASE_URL = '/api';
let currentView = 'todos';
let allCertificates = [];
let allTags = [];

// ==================== 工具函数 ====================

/**
 * 显示 Toast 通知
 */
function showToast(message, type = 'success') {
  const container = document.getElementById('toast-container');
  const toast = document.createElement('div');
  
  const bgColors = {
    success: 'bg-green-500',
    error: 'bg-red-500',
    warning: 'bg-yellow-500',
    info: 'bg-blue-500'
  };
  
  const icons = {
    success: 'fa-check-circle',
    error: 'fa-exclamation-circle',
    warning: 'fa-exclamation-triangle',
    info: 'fa-info-circle'
  };
  
  toast.className = `toast ${bgColors[type]} text-white px-6 py-4 rounded-xl shadow-2xl flex items-center space-x-3`;
  toast.innerHTML = `
    <i class="fas ${icons[type]} text-xl"></i>
    <span class="font-medium">${message}</span>
  `;
  
  container.appendChild(toast);
  
  // 3秒后自动移除
  setTimeout(() => {
    toast.style.animation = 'slideIn 0.3s ease reverse';
    setTimeout(() => {
      container.removeChild(toast);
    }, 300);
  }, 3000);
}

/**
 * 显示错误信息
 */
function showError(elementId, message) {
  const errorDiv = document.getElementById(elementId);
  const messageSpan = document.getElementById(`${elementId}-message`);
  
  if (errorDiv && messageSpan) {
    messageSpan.textContent = message;
    errorDiv.classList.remove('hidden');
    
    setTimeout(() => {
      errorDiv.classList.add('hidden');
    }, 5000);
  }
}

/**
 * 计算剩余天数
 */
function calculateDaysRemaining(expiryDate) {
  const now = new Date();
  const expiry = new Date(expiryDate);
  const diffTime = expiry.getTime() - now.getTime();
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));
  return diffDays;
}

/**
 * 格式化日期
 */
function formatDate(dateString) {
  if (!dateString) return '未知';
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
}

/**
 * 获取证书状态的样式类
 */
function getCertStatusClass(status, daysRemaining) {
  if (status === 'error') {
    return {
      bg: 'bg-gradient-to-br from-gray-50 to-gray-100',
      border: 'border-gray-300',
      text: 'text-gray-600',
      badge: 'bg-gray-500'
    };
  }
  
  if (status === 'expired' || daysRemaining < 0) {
    return {
      bg: 'bg-gradient-to-br from-red-50 to-pink-100',
      border: 'border-red-300',
      text: 'text-red-600',
      badge: 'bg-red-500'
    };
  }
  
  if (daysRemaining <= 7) {
    return {
      bg: 'bg-gradient-to-br from-yellow-50 to-orange-100',
      border: 'border-yellow-300',
      text: 'text-yellow-700',
      badge: 'bg-yellow-500'
    };
  }
  
  if (daysRemaining <= 30) {
    return {
      bg: 'bg-gradient-to-br from-yellow-50 to-yellow-100',
      border: 'border-yellow-200',
      text: 'text-yellow-600',
      badge: 'bg-yellow-400'
    };
  }
  
  return {
    bg: 'bg-gradient-to-br from-green-50 to-emerald-100',
    border: 'border-green-300',
    text: 'text-green-600',
    badge: 'bg-green-500'
  };
}

// ==================== 视图切换 ====================

/**
 * 切换视图
 */
function switchTab(view) {
  currentView = view;
  
  // 更新标签按钮状态
  document.querySelectorAll('.tab-button').forEach(btn => {
    btn.classList.remove('tab-active');
  });
  document.getElementById(`tab-${view}`).classList.add('tab-active');
  
  // 切换视图
  document.querySelectorAll('.view-container').forEach(container => {
    container.classList.add('hidden');
  });
  document.getElementById(`view-${view}`).classList.remove('hidden');
  
  // 加载数据
  if (view === 'todos') {
    fetchTodos();
  } else if (view === 'certificates') {
    fetchCertificates();
    loadTags();
  }
}

// ==================== 待办事项功能 ====================

/**
 * 获取所有待办事项
 */
async function fetchTodos() {
  const loadingDiv = document.getElementById('todo-loading');
  const listDiv = document.getElementById('todo-list');
  const emptyDiv = document.getElementById('todo-empty');
  const countSpan = document.getElementById('todo-count');
  
  loadingDiv.classList.remove('hidden');
  listDiv.classList.add('hidden');
  emptyDiv.classList.add('hidden');
  
  try {
    const response = await fetch(`${API_BASE_URL}/todos`);
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '获取待办事项失败');
    }
    
    loadingDiv.classList.add('hidden');
    
    if (result.data.length === 0) {
      emptyDiv.classList.remove('hidden');
      countSpan.textContent = '0';
    } else {
      renderTodos(result.data);
      listDiv.classList.remove('hidden');
      countSpan.textContent = result.data.length;
    }
    
  } catch (error) {
    loadingDiv.classList.add('hidden');
    showError('todo-error', `获取待办事项失败: ${error.message}`);
    showToast(`获取待办事项失败: ${error.message}`, 'error');
  }
}

/**
 * 渲染待办事项列表
 */
function renderTodos(todos) {
  const listDiv = document.getElementById('todo-list');
  listDiv.innerHTML = '';
  
  todos.forEach(todo => {
    const li = document.createElement('li');
    li.className = 'bg-white rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow duration-200 flex items-center justify-between fade-in';
    
    li.innerHTML = `
      <div class="flex items-center flex-1 min-w-0">
        <input 
          type="checkbox" 
          ${todo.completed ? 'checked' : ''} 
          onchange="toggleTodo(${todo.id}, ${!todo.completed})"
          class="w-5 h-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500 cursor-pointer flex-shrink-0">
        <span class="ml-3 flex-1 ${todo.completed ? 'line-through text-gray-400' : 'text-gray-800'} truncate">
          ${escapeHtml(todo.title)}
        </span>
      </div>
      <div class="flex items-center space-x-2 ml-4 flex-shrink-0">
        <button 
          onclick="editTodo(${todo.id}, '${escapeHtml(todo.title)}', ${todo.completed})" 
          class="px-3 py-1.5 bg-blue-100 text-blue-600 rounded-lg hover:bg-blue-200 transition-colors text-sm">
          <i class="fas fa-edit"></i>
        </button>
        <button 
          onclick="deleteTodo(${todo.id})" 
          class="px-3 py-1.5 bg-red-100 text-red-600 rounded-lg hover:bg-red-200 transition-colors text-sm">
          <i class="fas fa-trash"></i>
        </button>
      </div>
    `;
    
    listDiv.appendChild(li);
  });
}

/**
 * 添加待办事项
 */
async function addTodo() {
  const input = document.getElementById('new-todo-input');
  const title = input.value.trim();
  
  if (!title) {
    showToast('待办事项不能为空', 'warning');
    return;
  }
  
  try {
    const response = await fetch(`${API_BASE_URL}/todos`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ title, completed: false })
    });
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '创建待办事项失败');
    }
    
    input.value = '';
    showToast('添加成功', 'success');
    fetchTodos();
    
  } catch (error) {
    showToast(`添加失败: ${error.message}`, 'error');
  }
}

/**
 * 切换待办事项状态
 */
async function toggleTodo(id, completed) {
  try {
    const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ completed })
    });
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '更新待办事项失败');
    }
    
    fetchTodos();
    
  } catch (error) {
    showToast(`更新失败: ${error.message}`, 'error');
  }
}

/**
 * 编辑待办事项
 */
async function editTodo(id, currentTitle, completed) {
  const newTitle = prompt('编辑待办事项', currentTitle);
  
  if (newTitle && newTitle.trim() !== '' && newTitle !== currentTitle) {
    try {
      const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ title: newTitle.trim(), completed })
      });
      
      if (!response.ok) {
        throw new Error(`HTTP错误 ${response.status}`);
      }
      
      const result = await response.json();
      
      if (!result.success) {
        throw new Error(result.error || '更新待办事项失败');
      }
      
      showToast('更新成功', 'success');
      fetchTodos();
      
    } catch (error) {
      showToast(`更新失败: ${error.message}`, 'error');
    }
  }
}

/**
 * 删除待办事项
 */
async function deleteTodo(id) {
  if (!confirm('确定要删除这个待办事项吗？')) {
    return;
  }
  
  try {
    const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
      method: 'DELETE'
    });
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '删除待办事项失败');
    }
    
    showToast('删除成功', 'success');
    fetchTodos();
    
  } catch (error) {
    showToast(`删除失败: ${error.message}`, 'error');
  }
}

// ==================== SSL 证书监控功能 ====================

/**
 * 获取所有证书
 */
async function fetchCertificates() {
  const loadingDiv = document.getElementById('cert-loading');
  const gridDiv = document.getElementById('cert-grid');
  const emptyDiv = document.getElementById('cert-empty');
  
  loadingDiv.classList.remove('hidden');
  gridDiv.classList.add('hidden');
  emptyDiv.classList.add('hidden');
  
  try {
    const response = await fetch(`${API_BASE_URL}/certificates`);
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '获取证书列表失败');
    }
    
    allCertificates = result.data;
    loadingDiv.classList.add('hidden');
    
    if (result.data.length === 0) {
      emptyDiv.classList.remove('hidden');
    } else {
      renderCertificates(result.data);
      gridDiv.classList.remove('hidden');
    }
    
    // 启动倒计时更新
    startCountdownTimer();
    
  } catch (error) {
    loadingDiv.classList.add('hidden');
    showError('cert-error', `获取证书列表失败: ${error.message}`);
    showToast(`获取证书列表失败: ${error.message}`, 'error');
  }
}

/**
 * 渲染证书列表
 */
function renderCertificates(certificates) {
  const gridDiv = document.getElementById('cert-grid');
  gridDiv.innerHTML = '';
  
  certificates.forEach(cert => {
    const daysRemaining = cert.expiry_date ? calculateDaysRemaining(cert.expiry_date) : null;
    const statusClass = getCertStatusClass(cert.status, daysRemaining);
    
    const card = document.createElement('div');
    card.className = `${statusClass.bg} border-2 ${statusClass.border} rounded-2xl p-6 shadow-lg card-hover fade-in`;
    card.setAttribute('data-cert-id', cert.id);
    
    // 状态标识
    let statusBadge = '';
    if (cert.status === 'pending') {
      statusBadge = '<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-500 text-white"><i class="fas fa-clock mr-1"></i>待检查</span>';
    } else if (cert.status === 'error') {
      statusBadge = '<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-500 text-white"><i class="fas fa-exclamation-circle mr-1"></i>检查失败</span>';
    } else if (daysRemaining !== null) {
      if (daysRemaining < 0) {
        statusBadge = `<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusClass.badge} text-white"><i class="fas fa-times-circle mr-1"></i>已过期</span>`;
      } else if (daysRemaining <= 7) {
        statusBadge = `<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusClass.badge} text-white"><i class="fas fa-exclamation-triangle mr-1"></i>即将过期</span>`;
      } else {
        statusBadge = `<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusClass.badge} text-white"><i class="fas fa-check-circle mr-1"></i>正常</span>`;
      }
    }
    
    card.innerHTML = `
      <div class="flex items-start justify-between mb-4">
        <div class="flex items-center space-x-2 flex-1 min-w-0">
          <i class="fas fa-globe ${statusClass.text} text-xl flex-shrink-0"></i>
          <h3 class="font-bold text-gray-800 truncate">${escapeHtml(cert.domain)}</h3>
        </div>
        <button 
          onclick="deleteCertificate(${cert.id})" 
          class="text-gray-400 hover:text-red-500 transition-colors ml-2 flex-shrink-0">
          <i class="fas fa-times text-lg"></i>
        </button>
      </div>
      
      <div class="text-center py-6 mb-4">
        ${daysRemaining !== null && cert.status !== 'error' ? `
          <div class="text-5xl font-bold ${statusClass.text} mb-2" data-countdown="${cert.expiry_date}">
            ${daysRemaining >= 0 ? daysRemaining : 0}
          </div>
          <div class="text-sm text-gray-600 font-medium">
            ${daysRemaining >= 0 ? '天后到期' : '已过期'}
          </div>
        ` : `
          <div class="text-3xl font-bold text-gray-400 mb-2">
            <i class="fas fa-question-circle"></i>
          </div>
          <div class="text-sm text-gray-600 font-medium">
            ${cert.status === 'pending' ? '正在检查中...' : '检查失败'}
          </div>
        `}
      </div>
      
      <div class="space-y-2 text-sm mb-4">
        ${cert.expiry_date ? `
          <div class="flex items-center text-gray-600">
            <i class="fas fa-calendar-alt w-5 flex-shrink-0"></i>
            <span class="ml-2 truncate">${formatDate(cert.expiry_date)}</span>
          </div>
        ` : ''}
        ${cert.issuer ? `
          <div class="flex items-center text-gray-600">
            <i class="fas fa-certificate w-5 flex-shrink-0"></i>
            <span class="ml-2 truncate">${escapeHtml(cert.issuer)}</span>
          </div>
        ` : ''}
        ${cert.error_message ? `
          <div class="flex items-start text-red-600">
            <i class="fas fa-exclamation-triangle w-5 flex-shrink-0 mt-0.5"></i>
            <span class="ml-2 text-xs">${escapeHtml(cert.error_message)}</span>
          </div>
        ` : ''}
      </div>
      
      <div class="flex flex-wrap gap-1.5 mb-3">
        ${cert.tags && cert.tags.length > 0 ? cert.tags.map(tag => `
          <span class="tag-badge inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-700">
            <i class="fas fa-tag mr-1"></i>
            ${escapeHtml(tag)}
          </span>
        `).join('') : '<span class="text-xs text-gray-400"><i class="fas fa-tag mr-1"></i>无标签</span>'}
      </div>
      
      <div class="flex gap-2 pt-3 border-t border-gray-200">
        ${statusBadge}
        <button 
          onclick="checkSingleCertificate(${cert.id})" 
          class="flex-1 px-3 py-1.5 bg-white hover:bg-gray-50 text-gray-700 text-xs font-medium rounded-lg border border-gray-300 transition-colors">
          <i class="fas fa-sync-alt mr-1"></i>
          刷新
        </button>
        <button 
          onclick="editCertificateTags(${cert.id}, ${JSON.stringify(cert.tags || []).replace(/"/g, '&quot;')})" 
          class="flex-1 px-3 py-1.5 bg-white hover:bg-gray-50 text-gray-700 text-xs font-medium rounded-lg border border-gray-300 transition-colors">
          <i class="fas fa-tags mr-1"></i>
          标签
        </button>
      </div>
    `;
    
    gridDiv.appendChild(card);
  });
}

/**
 * 启动倒计时定时器（每60秒更新一次）
 */
let countdownInterval = null;

function startCountdownTimer() {
  if (countdownInterval) {
    clearInterval(countdownInterval);
  }
  
  countdownInterval = setInterval(() => {
    if (currentView === 'certificates') {
      document.querySelectorAll('[data-countdown]').forEach(elem => {
        const expiryDate = elem.getAttribute('data-countdown');
        const days = calculateDaysRemaining(expiryDate);
        elem.textContent = days >= 0 ? days : 0;
      });
    }
  }, 60000); // 每60秒更新一次
}

/**
 * 显示添加证书模态框
 */
function showAddCertificateModal() {
  document.getElementById('add-cert-modal').classList.remove('hidden');
  document.getElementById('new-cert-domain').value = '';
  loadTagCheckboxes();
}

/**
 * 关闭添加证书模态框
 */
function closeAddCertificateModal() {
  document.getElementById('add-cert-modal').classList.add('hidden');
}

/**
 * 加载标签复选框
 */
function loadTagCheckboxes() {
  const container = document.getElementById('tag-checkboxes');
  
  if (allTags.length === 0) {
    container.innerHTML = '<p class="text-sm text-gray-400">还没有标签，请先创建</p>';
    return;
  }
  
  container.innerHTML = allTags.map(tag => `
    <label class="flex items-center space-x-2 cursor-pointer hover:bg-gray-50 p-2 rounded-lg transition-colors">
      <input type="checkbox" value="${escapeHtml(tag)}" class="tag-checkbox w-4 h-4 text-primary-600 rounded focus:ring-2 focus:ring-primary-500">
      <span class="text-sm text-gray-700">${escapeHtml(tag)}</span>
    </label>
  `).join('');
}

/**
 * 添加新标签
 */
function addNewTag() {
  const input = document.getElementById('new-tag-input');
  const tag = input.value.trim();
  
  if (!tag) {
    showToast('标签名称不能为空', 'warning');
    return;
  }
  
  if (allTags.includes(tag)) {
    showToast('标签已存在', 'warning');
    return;
  }
  
  allTags.push(tag);
  allTags.sort();
  input.value = '';
  loadTagCheckboxes();
  updateTagFilter();
  showToast('标签创建成功', 'success');
}

/**
 * 加载所有标签
 */
async function loadTags() {
  try {
    const response = await fetch(`${API_BASE_URL}/tags`);
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (result.success) {
      allTags = result.data || [];
      updateTagFilter();
    }
    
  } catch (error) {
    console.error('加载标签失败:', error);
  }
}

/**
 * 更新标签筛选下拉框
 */
function updateTagFilter() {
  const select = document.getElementById('tag-filter');
  const currentValue = select.value;
  
  select.innerHTML = '<option value="">全部标签</option>' + 
    allTags.map(tag => `<option value="${escapeHtml(tag)}">${escapeHtml(tag)}</option>`).join('');
  
  select.value = currentValue;
}

/**
 * 按标签筛选证书
 */
function filterCertificatesByTag() {
  const selectedTag = document.getElementById('tag-filter').value;
  
  if (!selectedTag) {
    renderCertificates(allCertificates);
  } else {
    const filtered = allCertificates.filter(cert => 
      cert.tags && cert.tags.includes(selectedTag)
    );
    renderCertificates(filtered);
  }
}

/**
 * 添加证书
 */
async function addCertificate() {
  const domainInput = document.getElementById('new-cert-domain');
  const domain = domainInput.value.trim().toLowerCase();
  
  if (!domain) {
    showToast('域名不能为空', 'warning');
    return;
  }
  
  // 获取选中的标签
  const selectedTags = Array.from(document.querySelectorAll('.tag-checkbox:checked'))
    .map(checkbox => checkbox.value);
  
  try {
    const response = await fetch(`${API_BASE_URL}/certificates`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ domain, tags: selectedTags })
    });
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '添加域名失败');
    }
    
    closeAddCertificateModal();
    showToast('域名添加成功，正在检查证书...', 'success');
    
    // 等待一会再刷新，让后台有时间检查
    setTimeout(() => {
      fetchCertificates();
    }, 2000);
    
  } catch (error) {
    showToast(`添加失败: ${error.message}`, 'error');
  }
}

/**
 * 删除证书
 */
async function deleteCertificate(id) {
  if (!confirm('确定要删除这个域名吗？')) {
    return;
  }
  
  try {
    const response = await fetch(`${API_BASE_URL}/certificates/${id}`, {
      method: 'DELETE'
    });
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '删除域名失败');
    }
    
    showToast('删除成功', 'success');
    fetchCertificates();
    
  } catch (error) {
    showToast(`删除失败: ${error.message}`, 'error');
  }
}

/**
 * 检查单个证书
 */
async function checkSingleCertificate(id) {
  try {
    const response = await fetch(`${API_BASE_URL}/certificates/${id}/check`, {
      method: 'POST'
    });
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '检查失败');
    }
    
    showToast('正在检查证书...', 'info');
    
    // 5秒后刷新
    setTimeout(() => {
      fetchCertificates();
    }, 5000);
    
  } catch (error) {
    showToast(`检查失败: ${error.message}`, 'error');
  }
}

/**
 * 检查所有证书
 */
async function checkAllCertificates() {
  try {
    const response = await fetch(`${API_BASE_URL}/certificates/check-all`, {
      method: 'POST'
    });
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '批量检查失败');
    }
    
    showToast(result.message || '正在检查所有证书...', 'info');
    
    // 10秒后刷新
    setTimeout(() => {
      fetchCertificates();
    }, 10000);
    
  } catch (error) {
    showToast(`批量检查失败: ${error.message}`, 'error');
  }
}

/**
 * 编辑证书标签
 */
async function editCertificateTags(id, currentTags) {
  await loadTags();
  
  // 创建临时模态框
  const modal = document.createElement('div');
  modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4';
  modal.innerHTML = `
    <div class="bg-white rounded-2xl shadow-2xl max-w-md w-full p-6 fade-in">
      <div class="flex items-center justify-between mb-6">
        <h3 class="text-xl font-bold text-gray-800">
          <i class="fas fa-tags text-primary-500 mr-2"></i>
          编辑标签
        </h3>
        <button onclick="this.closest('.fixed').remove()" class="text-gray-400 hover:text-gray-600 transition-colors">
          <i class="fas fa-times text-xl"></i>
        </button>
      </div>
      <div id="edit-tag-checkboxes" class="space-y-2 max-h-64 overflow-y-auto hide-scrollbar mb-4">
        ${allTags.map(tag => `
          <label class="flex items-center space-x-2 cursor-pointer hover:bg-gray-50 p-2 rounded-lg transition-colors">
            <input type="checkbox" value="${escapeHtml(tag)}" ${currentTags.includes(tag) ? 'checked' : ''} class="edit-tag-checkbox w-4 h-4 text-primary-600 rounded focus:ring-2 focus:ring-primary-500">
            <span class="text-sm text-gray-700">${escapeHtml(tag)}</span>
          </label>
        `).join('')}
      </div>
      <div class="flex gap-3">
        <button onclick="this.closest('.fixed').remove()" class="flex-1 px-4 py-3 bg-gray-100 text-gray-700 font-medium rounded-xl hover:bg-gray-200 transition-colors">
          取消
        </button>
        <button onclick="saveEditedTags(${id})" class="flex-1 px-4 py-3 gradient-bg text-white font-medium rounded-xl hover:shadow-lg hover:scale-105 transition-all duration-200">
          保存
        </button>
      </div>
    </div>
  `;
  
  document.body.appendChild(modal);
}

/**
 * 保存编辑的标签
 */
async function saveEditedTags(id) {
  const selectedTags = Array.from(document.querySelectorAll('.edit-tag-checkbox:checked'))
    .map(checkbox => checkbox.value);
  
  try {
    const response = await fetch(`${API_BASE_URL}/certificates/${id}/tags`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ tags: selectedTags })
    });
    
    if (!response.ok) {
      throw new Error(`HTTP错误 ${response.status}`);
    }
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error || '更新标签失败');
    }
    
    // 关闭模态框
    document.querySelector('.fixed.inset-0').remove();
    
    showToast('标签更新成功', 'success');
    fetchCertificates();
    
  } catch (error) {
    showToast(`更新失败: ${error.message}`, 'error');
  }
}

/**
 * HTML 转义
 */
function escapeHtml(text) {
  const map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  };
  return String(text).replace(/[&<>"']/g, m => map[m]);
}

// ==================== 初始化 ====================

document.addEventListener('DOMContentLoaded', () => {
  // 加载待办事项
  fetchTodos();
  
  // 模态框外部点击关闭
  document.getElementById('add-cert-modal').addEventListener('click', (e) => {
    if (e.target.id === 'add-cert-modal') {
      closeAddCertificateModal();
    }
  });
});

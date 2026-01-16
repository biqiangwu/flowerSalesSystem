// ==================== API 客户端 ====================
class ApiClient {
    constructor() {
        this.baseURL = '';
        this.sessionToken = this.getCookie('session_token') || null;
    }

    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
        return null;
    }

    setCookie(name, value, days = 1) {
        const date = new Date();
        date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
        document.cookie = `${name}=${value};expires=${date.toUTCString()};path=/`;
    }

    deleteCookie(name) {
        document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/`;
    }

    async request(url, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        const config = {
            ...options,
            headers
        };

        const response = await fetch(this.baseURL + url, config);

        if (response.status === 204) {
            return null;
        }

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || data.details || '请求失败');
        }

        return data;
    }

    // 认证相关
    async register(username, password) {
        const result = await this.request('/api/register', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        });
        return result.user;
    }

    async login(username, password) {
        const result = await this.request('/api/login', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        });
        this.sessionToken = result.token;
        this.setCookie('session_token', result.token);
        return result.user;
    }

    async logout() {
        try {
            await this.request('/api/logout', { method: 'POST' });
        } finally {
            this.sessionToken = null;
            this.deleteCookie('session_token');
        }
    }

    // 鲜花相关
    async getFlowers(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/api/flowers?${queryString}`);
    }

    async getFlower(sku) {
        return this.request(`/api/flowers/${sku}`);
    }

    async createFlower(data) {
        return this.request('/api/flowers', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    async updateFlower(sku, data) {
        return this.request(`/api/flowers/${sku}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    async deleteFlower(sku) {
        return this.request(`/api/flowers/${sku}`, { method: 'DELETE' });
    }

    async addStock(sku, quantity) {
        return this.request(`/api/flowers/${sku}/stock`, {
            method: 'POST',
            body: JSON.stringify({ quantity })
        });
    }

    // 地址相关
    async getAddresses() {
        return this.request('/api/addresses');
    }

    async createAddress(data) {
        return this.request('/api/addresses', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    async updateAddress(id, data) {
        return this.request(`/api/addresses/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    async deleteAddress(id) {
        return this.request(`/api/addresses/${id}`, { method: 'DELETE' });
    }

    // 订单相关
    async getOrders(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/api/orders?${queryString}`);
    }

    async getOrder(id) {
        return this.request(`/api/orders/${id}`);
    }

    async createOrder(data) {
        return this.request('/api/orders', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    async completeOrder(id) {
        return this.request(`/api/orders/${id}/complete`, { method: 'PUT' });
    }

    async cancelOrder(id) {
        return this.request(`/api/orders/${id}/cancel`, { method: 'PUT' });
    }

    async getOrderLogs(id) {
        return this.request(`/api/orders/${id}/logs`);
    }

    // 用户管理（管理员）
    async getUsers(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/api/users?${queryString}`);
    }

    async deleteUser(id) {
        return this.request(`/api/users/${id}`, { method: 'DELETE' });
    }

    async resetPassword(id, newPassword) {
        return this.request(`/api/users/${id}/password`, {
            method: 'PUT',
            body: JSON.stringify({ new_password: newPassword })
        });
    }

    // 获取当前用户信息
    async getCurrentUser() {
        try {
            return await this.request('/api/me');
        } catch (error) {
            return null;
        }
    }
}

// ==================== 状态管理 ====================
class Store {
    constructor() {
        this.user = null;
        this.cart = [];
        this.listeners = [];
    }

    subscribe(listener) {
        this.listeners.push(listener);
        return () => {
            this.listeners = this.listeners.filter(l => l !== listener);
        };
    }

    notify() {
        this.listeners.forEach(listener => listener(this));
    }

    setUser(user) {
        this.user = user;
        this.notify();
    }

    clearUser() {
        this.user = null;
        this.notify();
    }

    addToCart(flower, quantity = 1) {
        const existingItem = this.cart.find(item => item.sku === flower.sku);
        if (existingItem) {
            existingItem.quantity += quantity;
        } else {
            this.cart.push({
                sku: flower.sku,
                name: flower.name,
                price: flower.sale_price,
                quantity: quantity
            });
        }
        this.notify();
    }

    removeFromCart(sku) {
        this.cart = this.cart.filter(item => item.sku !== sku);
        this.notify();
    }

    updateCartQuantity(sku, quantity) {
        const item = this.cart.find(item => item.sku === sku);
        if (item) {
            item.quantity = quantity;
            if (item.quantity <= 0) {
                this.removeFromCart(sku);
            } else {
                this.notify();
            }
        }
    }

    clearCart() {
        this.cart = [];
        this.notify();
    }

    getCartTotal() {
        return this.cart.reduce((total, item) => total + (item.price * item.quantity), 0);
    }
}

// ==================== 路由管理 ====================
class Router {
    constructor() {
        this.routes = {};
        this.currentRoute = null;
        window.addEventListener('hashchange', () => this.handleRoute());
    }

    register(path, handler) {
        this.routes[path] = handler;
    }

    navigate(path) {
        window.location.hash = path;
    }

    async handleRoute() {
        const hash = window.location.hash.slice(1) || '/';
        this.currentRoute = hash;

        // 查找匹配的路由
        let handler = this.routes[hash];
        if (!handler) {
            // 尝试匹配动态路由
            for (const route in this.routes) {
                const pattern = route.replace(/:([^/]+)/g, '([^/]+)');
                const regex = new RegExp(`^${pattern}$`);
                const match = hash.match(regex);
                if (match) {
                    const params = route.split('/').filter(p => p.startsWith(':'))
                        .map((p, i) => [p.slice(1), match[i + 1]])
                        .reduce((obj, [key, val]) => ({ ...obj, [key]: val }), {});
                    handler = this.routes[route];
                    if (handler) {
                        return handler(params);
                    }
                    break;
                }
            }
        }

        if (handler) {
            await handler();
        } else {
            this.routes['/']?.();
        }
    }

    start() {
        this.handleRoute();
    }
}

// ==================== UI 工具 ====================
class UI {
    static showLoading() {
        document.getElementById('loadingOverlay').style.display = 'flex';
    }

    static hideLoading() {
        document.getElementById('loadingOverlay').style.display = 'none';
    }

    static toast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.innerHTML = `
            <span class="toast-message">${message}</span>
            <button class="toast-close">&times;</button>
        `;
        container.appendChild(toast);

        const closeBtn = toast.querySelector('.toast-close');
        closeBtn.onclick = () => toast.remove();

        setTimeout(() => toast.remove(), 3000);
    }

    static renderPage(content) {
        document.getElementById('pageContent').innerHTML = content;
    }

    static updateNav(user) {
        const navMenu = document.getElementById('navMenu');
        const userInfo = document.getElementById('userInfo');
        const logoutBtn = document.getElementById('logoutBtn');

        if (user) {
            userInfo.textContent = `${user.username} (${this.getRoleName(user.role)})`;
            logoutBtn.style.display = 'inline-block';

            // 根据角色显示不同的菜单项
            let menuHtml = '<li><a href="#/" class="nav-link">首页</a></li>';
            menuHtml += '<li><a href="#/flowers" class="nav-link">鲜花列表</a></li>';
            menuHtml += '<li><a href="#/orders" class="nav-link">我的订单</a></li>';

            if (user.role === 'clerk' || user.role === 'admin') {
                menuHtml += '<li><a href="#/orders/all" class="nav-link">订单管理</a></li>';
            }

            if (user.role === 'admin') {
                menuHtml += '<li><a href="#/users" class="nav-link">用户管理</a></li>';
            }

            navMenu.innerHTML = menuHtml;
        } else {
            userInfo.textContent = '';
            logoutBtn.style.display = 'none';
            navMenu.innerHTML = '<li><a href="#/" class="nav-link">首页</a></li>';
        }

        // 更新当前激活的导航项
        const currentPath = window.location.hash.slice(1) || '/';
        navMenu.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('href') === `#${currentPath}`) {
                link.classList.add('active');
            }
        });
    }

    static getRoleName(role) {
        const roleNames = {
            'customer': '顾客',
            'clerk': '店员',
            'admin': '管理员'
        };
        return roleNames[role] || role;
    }
}

// ==================== 页面渲染器 ====================
class Pages {
    // 首页
    static home() {
        UI.renderPage(`
            <div class="card">
                <div class="card-body text-center">
                    <h1 style="color: var(--primary-color); margin-bottom: 20px;">欢迎来到鲜花销售系统</h1>
                    <p class="text-muted mb-2">精选鲜花，为您传递美好</p>
                    <div class="mt-3">
                        <button class="btn btn-primary" onclick="router.navigate('/flowers')">浏览鲜花</button>
                    </div>
                </div>
            </div>
        `);
    }

    // 登录/注册页面
    static auth() {
        UI.renderPage(`
            <div class="auth-container">
                <div class="card">
                    <div class="auth-tabs">
                        <div class="auth-tab active" data-tab="login" onclick="App.switchAuthTab('login')">登录</div>
                        <div class="auth-tab" data-tab="register" onclick="App.switchAuthTab('register')">注册</div>
                    </div>
                    <div class="card-body">
                        <form id="loginForm">
                            <div class="form-group">
                                <label class="form-label">用户名</label>
                                <input type="text" class="form-control" name="username" required>
                            </div>
                            <div class="form-group">
                                <label class="form-label">密码</label>
                                <input type="password" class="form-control" name="password" required>
                            </div>
                            <button type="submit" class="btn btn-primary w-100">登录</button>
                        </form>
                        <form id="registerForm" style="display: none;">
                            <div class="form-group">
                                <label class="form-label">用户名</label>
                                <input type="text" class="form-control" name="username" required>
                            </div>
                            <div class="form-group">
                                <label class="form-label">密码</label>
                                <input type="password" class="form-control" name="password" required>
                            </div>
                            <div class="form-group">
                                <label class="form-label">确认密码</label>
                                <input type="password" class="form-control" name="confirmPassword" required>
                            </div>
                            <button type="submit" class="btn btn-primary w-100">注册</button>
                        </form>
                    </div>
                </div>
            </div>
        `);

        // 绑定表单事件
        document.getElementById('loginForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            try {
                UI.showLoading();
                const user = await api.login(formData.get('username'), formData.get('password'));
                store.setUser(user);
                UI.toast('登录成功', 'success');
                router.navigate('/flowers');
            } catch (error) {
                UI.toast(error.message, 'error');
            } finally {
                UI.hideLoading();
            }
        };

        document.getElementById('registerForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const password = formData.get('password');
            const confirmPassword = formData.get('confirmPassword');

            if (password !== confirmPassword) {
                UI.toast('两次输入的密码不一致', 'error');
                return;
            }

            try {
                UI.showLoading();
                await api.register(formData.get('username'), password);
                UI.toast('注册成功，请登录', 'success');
                App.switchAuthTab('login');
            } catch (error) {
                UI.toast(error.message, 'error');
            } finally {
                UI.hideLoading();
            }
        };
    }

    static switchAuthTab(tab) {
        document.querySelectorAll('.auth-tab').forEach(el => {
            el.classList.toggle('active', el.dataset.tab === tab);
        });
        document.getElementById('loginForm').style.display = tab === 'login' ? 'block' : 'none';
        document.getElementById('registerForm').style.display = tab === 'register' ? 'block' : 'none';
    }

    // 鲜花列表页面
    static async flowers() {
        try {
            UI.showLoading();
            const flowers = await api.getFlowers();

            const flowersHtml = flowers.map(flower => `
                <div class="product-card">
                    <div class="product-card-body">
                        <h3 class="product-title">${flower.name}</h3>
                        <p class="product-info">SKU: ${flower.sku}</p>
                        <p class="product-info">产地: ${flower.origin}</p>
                        <p class="product-info">花期: ${flower.shelf_life}</p>
                        <p class="product-price">¥${flower.sale_price.toFixed(2)}</p>
                        <span class="product-stock ${flower.stock < 10 ? 'low' : 'normal'}">
                            库存: ${flower.stock} 支
                        </span>
                        ${store.user && store.user.role === 'customer' ? `
                            <div class="mt-2">
                                <button class="btn btn-primary btn-sm w-100"
                                    onclick="App.addToCart('${flower.sku}')"
                                    ${flower.stock <= 0 ? 'disabled' : ''}>
                                    ${flower.stock > 0 ? '加入购物车' : '已售罄'}
                                </button>
                            </div>
                        ` : ''}
                    </div>
                </div>
            `).join('');

            const cartHtml = store.user && store.user.role === 'customer' ? `
                <div class="card" style="position: sticky; top: 80px;">
                    <div class="card-header">购物车</div>
                    <div class="card-body">
                        ${store.cart.length === 0 ? '<p class="text-muted text-center">购物车为空</p>' : ''}
                        <div id="cartItems">
                            ${store.cart.map(item => `
                                <div class="d-flex justify-content-between align-items-center mb-1">
                                    <span>${item.name} x ${item.quantity}</span>
                                    <span class="text-primary">¥${(item.price * item.quantity).toFixed(2)}</span>
                                </div>
                            `).join('')}
                        </div>
                        ${store.cart.length > 0 ? `
                            <hr class="mt-2 mb-2">
                            <div class="d-flex justify-content-between align-items-center">
                                <strong>总计:</strong>
                                <strong class="text-primary">¥${store.getCartTotal().toFixed(2)}</strong>
                            </div>
                            <button class="btn btn-primary w-100 mt-2" onclick="router.navigate('/checkout')">
                                去结算
                            </button>
                            <button class="btn btn-secondary w-100 mt-1" onclick="App.clearCart()">
                                清空
                            </button>
                        ` : ''}
                    </div>
                </div>
            ` : '';

            UI.renderPage(`
                <div class="row">
                    <div class="${cartHtml ? 'col-8' : 'col-12'}">
                        <div class="d-flex justify-content-between align-items-center mb-2">
                            <h2>鲜花列表</h2>
                        </div>
                        <div class="product-grid">
                            ${flowersHtml}
                        </div>
                    </div>
                    ${cartHtml ? '<div class="col-4">' + cartHtml + '</div>' : ''}
                </div>
            `);
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    // 订单列表页面
    static async orders(all = false) {
        try {
            UI.showLoading();
            const orders = await api.getOrders(all ? { all: 'true' } : {});

            const ordersHtml = orders.length === 0 ? `
                <p class="text-muted text-center">暂无订单</p>
            ` : orders.map(order => `
                <div class="card mb-2">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <span>订单号: ${order.order_no}</span>
                        <span class="order-status ${order.status}">${App.getStatusName(order.status)}</span>
                    </div>
                    <div class="card-body">
                        <p><strong>创建时间:</strong> ${new Date(order.created_at).toLocaleString()}</p>
                        <p><strong>总金额:</strong> <span class="text-primary">¥${order.total_amount.toFixed(2)}</span></p>
                        <div class="mt-2">
                            ${order.items.map(item => `
                                <div class="d-flex justify-content-between">
                                    <span>${item.flower_name} x ${item.quantity}</span>
                                    <span>¥${item.subtotal.toFixed(2)}</span>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                    ${all || (order.status === 'pending' && (!store.user || store.user.role === 'customer')) ? `
                        <div class="card-footer">
                            ${order.status === 'pending' ? `
                                ${!all ? '<button class="btn btn-danger btn-sm" onclick="App.cancelOrder(' + order.id + ')">取消订单</button>' : ''}
                                ${all && (store.user?.role === 'clerk' || store.user?.role === 'admin') ? `
                                    <button class="btn btn-success btn-sm" onclick="App.completeOrder(' + order.id + ')">完成订单</button>
                                    <button class="btn btn-danger btn-sm" onclick="App.cancelOrder(' + order.id + ')">取消订单</button>
                                ` : ''}
                            ` : ''}
                            ${all ? `<button class="btn btn-secondary btn-sm" onclick="App.viewOrderLogs(' + order.id + ')">查看日志</button>` : ''}
                        </div>
                    ` : ''}
                </div>
            `).join('');

            UI.renderPage(`
                <div class="container">
                    <h2 class="mb-2">${all ? '订单管理' : '我的订单'}</h2>
                    ${ordersHtml}
                </div>
            `);
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    // 用户管理页面（管理员）
    static async users() {
        if (!store.user || store.user.role !== 'admin') {
            UI.toast('无权限访问', 'error');
            router.navigate('/');
            return;
        }

        try {
            UI.showLoading();
            const users = await api.getUsers();

            const usersHtml = users.map(user => `
                <tr>
                    <td>${user.id}</td>
                    <td>${user.username}</td>
                    <td><span class="badge badge-${user.role === 'admin' ? 'danger' : user.role === 'clerk' ? 'warning' : 'info'}">${UI.getRoleName(user.role)}</span></td>
                    <td>${new Date(user.created_at).toLocaleDateString()}</td>
                    <td>
                        ${user.role !== 'admin' ? `
                            <button class="btn btn-danger btn-sm" onclick="App.deleteUser(${user.id})">删除</button>
                            <button class="btn btn-secondary btn-sm" onclick="App.resetPassword(${user.id})">重置密码</button>
                        ` : '-'}
                    </td>
                </tr>
            `).join('');

            UI.renderPage(`
                <div class="container">
                    <h2 class="mb-2">用户管理</h2>
                    <div class="card">
                        <div class="card-body">
                            <table class="table">
                                <thead>
                                    <tr>
                                        <th>ID</th>
                                        <th>用户名</th>
                                        <th>角色</th>
                                        <th>创建时间</th>
                                        <th>操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${usersHtml}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            `);
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    static getStatusName(status) {
        const names = {
            'pending': '待处理',
            'completed': '已完成',
            'cancelled': '已取消'
        };
        return names[status] || status;
    }

    // 添加到购物车
    static async addToCart(sku) {
        try {
            UI.showLoading();
            const flower = await api.getFlower(sku);
            store.addToCart(flower, 1);
            UI.toast('已加入购物车', 'success');
            router.handleRoute();
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    static clearCart() {
        store.clearCart();
        UI.toast('购物车已清空', 'info');
        router.handleRoute();
    }

    // 完成订单
    static async completeOrder(id) {
        if (!confirm('确认完成此订单？')) return;

        try {
            UI.showLoading();
            await api.completeOrder(id);
            UI.toast('订单已完成', 'success');
            router.handleRoute();
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    // 取消订单
    static async cancelOrder(id) {
        if (!confirm('确认取消此订单？')) return;

        try {
            UI.showLoading();
            await api.cancelOrder(id);
            UI.toast('订单已取消', 'success');
            router.handleRoute();
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    // 删除用户
    static async deleteUser(id) {
        if (!confirm('确认删除此用户？')) return;

        try {
            UI.showLoading();
            await api.deleteUser(id);
            UI.toast('用户已删除', 'success');
            router.handleRoute();
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    // 重置密码
    static async resetPassword(id) {
        const newPassword = prompt('请输入新密码:');
        if (!newPassword) return;

        try {
            UI.showLoading();
            await api.resetPassword(id, newPassword);
            UI.toast('密码已重置', 'success');
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    // 查看订单日志
    static async viewOrderLogs(id) {
        try {
            UI.showLoading();
            const logs = await api.getOrderLogs(id);

            const logsHtml = logs.map(log => `
                <tr>
                    <td>${new Date(log.created_at).toLocaleString()}</td>
                    <td>用户ID: ${log.operator_id}</td>
                    <td>${log.action}</td>
                    <td>${log.old_status || '-'}</td>
                    <td>${log.new_status}</td>
                </tr>
            `).join('');

            UI.renderPage(`
                <div class="container">
                    <h2 class="mb-2">订单日志</h2>
                    <button class="btn btn-secondary mb-2" onclick="router.navigate('/orders/all')">返回</button>
                    <div class="card">
                        <div class="card-body">
                            <table class="table">
                                <thead>
                                    <tr>
                                        <th>时间</th>
                                        <th>操作人</th>
                                        <th>操作</th>
                                        <th>变更前状态</th>
                                        <th>变更后状态</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${logsHtml}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            `);
        } catch (error) {
            UI.toast(error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }
}

// ==================== 应用初始化 ====================
const api = new ApiClient();
const store = new Store();
const router = new Router();

const App = {
    async init() {
        // 注册路由
        router.register('/', () => Pages.home());
        router.register('/auth', () => Pages.auth());
        router.register('/flowers', () => Pages.flowers());
        router.register('/orders', () => Pages.orders(false));
        router.register('/orders/all', () => Pages.orders(true));
        router.register('/users', () => Pages.users());

        // 订阅状态变化
        store.subscribe(() => {
            UI.updateNav(store.user);
        });

        // 获取当前用户
        const user = await api.getCurrentUser();
        if (user) {
            store.setUser(user);
        }

        // 绑定登出按钮
        document.getElementById('logoutBtn').onclick = async () => {
            try {
                UI.showLoading();
                await api.logout();
                store.clearUser();
                store.clearCart();
                UI.toast('已登出', 'success');
                router.navigate('/');
            } catch (error) {
                UI.toast(error.message, 'error');
            } finally {
                UI.hideLoading();
            }
        };

        // 启动路由
        router.start();
        UI.updateNav(store.user);
    }
};

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => App.init());

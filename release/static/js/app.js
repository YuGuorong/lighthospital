// 全局变量
let currentUser = null;
let currentPage = 'dashboard';

// API基础URL
const API_BASE = '/api';

// 初始化应用
document.addEventListener('DOMContentLoaded', function() {
    initApp();
});

// 初始化应用
function initApp() {
    // 检查登录状态
    checkAuthStatus();
    
    // 绑定事件
    bindEvents();
    
    // 加载页面内容
    loadPage('dashboard');
}

// 检查认证状态
async function checkAuthStatus() {
    try {
        const response = await fetch(`${API_BASE}/auth/current`);
        if (response.ok) {
            const data = await response.json();
            currentUser = data.user;
            showMainApp();
        } else {
            showLoginPage();
        }
    } catch (error) {
        console.error('检查认证状态失败:', error);
        showLoginPage();
    }
}

// 显示登录页面
function showLoginPage() {
    document.getElementById('loginPage').classList.remove('hidden');
    document.getElementById('mainApp').classList.add('hidden');
}

// 显示主应用
function showMainApp() {
    document.getElementById('loginPage').classList.add('hidden');
    document.getElementById('mainApp').classList.remove('hidden');
    document.getElementById('userInfo').textContent = `${currentUser.name} (${currentUser.role})`;
}

// 绑定事件
function bindEvents() {
    // 登录表单
    document.getElementById('loginForm').addEventListener('submit', handleLogin);
    
    // 退出登录
    document.getElementById('logoutBtn').addEventListener('click', handleLogout);
    
    // 侧边栏导航
    document.querySelectorAll('.nav-link[data-page]').forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const page = this.getAttribute('data-page');
            loadPage(page);
        });
    });
    
    // 搜索功能
    document.getElementById('patientSearch').addEventListener('input', debounce(loadPatients, 500));
    document.getElementById('medicineSearch').addEventListener('input', debounce(loadMedicines, 500));
    document.getElementById('prescriptionSearch').addEventListener('input', debounce(loadPrescriptions, 500));
    document.getElementById('appointmentSearch').addEventListener('input', debounce(loadAppointments, 500));
    
    // 筛选功能
    document.getElementById('medicineCategory').addEventListener('change', loadMedicines);
    document.getElementById('prescriptionStatus').addEventListener('change', loadPrescriptions);
    document.getElementById('appointmentDate').addEventListener('change', loadAppointments);
}

// 处理登录
async function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    
    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password }),
        });
        
        if (response.ok) {
            const data = await response.json();
            currentUser = data.user;
            showMainApp();
            loadPage('dashboard');
        } else {
            const error = await response.json();
            alert(error.error || '登录失败');
        }
    } catch (error) {
        console.error('登录失败:', error);
        alert('登录失败，请检查网络连接');
    }
}

// 处理退出登录
async function handleLogout() {
    try {
        await fetch(`${API_BASE}/auth/logout`, { method: 'POST' });
        currentUser = null;
        showLoginPage();
    } catch (error) {
        console.error('退出登录失败:', error);
    }
}

// 加载页面
function loadPage(page) {
    // 更新当前页面
    currentPage = page;
    
    // 更新导航状态
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    document.querySelector(`[data-page="${page}"]`).classList.add('active');
    
    // 隐藏所有页面
    document.querySelectorAll('.page-content').forEach(content => {
        content.classList.add('hidden');
    });
    
    // 显示目标页面
    document.getElementById(`${page}Page`).classList.remove('hidden');
    
    // 更新页面标题
    const titles = {
        dashboard: '仪表板',
        patients: '患者管理',
        medicines: '药品管理',
        prescriptions: '处方管理',
        appointments: '预约管理',
        reports: '统计报表'
    };
    document.getElementById('pageTitle').textContent = titles[page] || '页面';
    
    // 加载页面数据
    switch (page) {
        case 'dashboard':
            loadDashboard();
            break;
        case 'patients':
            loadPatients();
            break;
        case 'medicines':
            loadMedicines();
            break;
        case 'prescriptions':
            loadPrescriptions();
            break;
        case 'appointments':
            loadAppointments();
            break;
        case 'reports':
            loadReports();
            break;
    }
}

// 加载仪表板
async function loadDashboard() {
    try {
        // 加载统计数据
        const [patientsRes, medicinesRes, prescriptionsRes, appointmentsRes] = await Promise.all([
            fetch(`${API_BASE}/patients?limit=1`),
            fetch(`${API_BASE}/medicines?limit=1`),
            fetch(`${API_BASE}/prescriptions?limit=1`),
            fetch(`${API_BASE}/appointments/today`)
        ]);
        
        if (patientsRes.ok) {
            const data = await patientsRes.json();
            document.getElementById('patientCount').textContent = data.total;
        }
        
        if (medicinesRes.ok) {
            const data = await medicinesRes.json();
            document.getElementById('medicineCount').textContent = data.total;
        }
        
        if (prescriptionsRes.ok) {
            const data = await prescriptionsRes.json();
            document.getElementById('prescriptionCount').textContent = data.total;
        }
        
        if (appointmentsRes.ok) {
            const data = await appointmentsRes.json();
            document.getElementById('appointmentCount').textContent = data.appointments.length;
            displayTodayAppointments(data.appointments);
        }
        
        // 加载低库存药品
        const lowStockRes = await fetch(`${API_BASE}/medicines/low-stock`);
        if (lowStockRes.ok) {
            const data = await lowStockRes.json();
            displayLowStockMedicines(data.medicines);
        }
        
    } catch (error) {
        console.error('加载仪表板失败:', error);
    }
}

// 显示今日预约
function displayTodayAppointments(appointments) {
    const container = document.getElementById('todayAppointments');
    if (appointments.length === 0) {
        container.innerHTML = '<p class="text-muted">今日无预约</p>';
        return;
    }
    
    const html = appointments.map(apt => `
        <div class="d-flex justify-content-between align-items-center mb-2">
            <div>
                <strong>${apt.patient.name}</strong>
                <br><small class="text-muted">${apt.appointment_time}</small>
            </div>
            <span class="badge bg-primary">${apt.status}</span>
        </div>
    `).join('');
    
    container.innerHTML = html;
}

// 显示低库存药品
function displayLowStockMedicines(medicines) {
    const container = document.getElementById('lowStockMedicines');
    if (medicines.length === 0) {
        container.innerHTML = '<p class="text-muted">无低库存药品</p>';
        return;
    }
    
    const html = medicines.map(med => `
        <div class="d-flex justify-content-between align-items-center mb-2">
            <div>
                <strong>${med.name}</strong>
                <br><small class="text-muted">库存: ${med.stock}</small>
            </div>
            <span class="badge bg-warning">低库存</span>
        </div>
    `).join('');
    
    container.innerHTML = html;
}

// 加载患者列表
async function loadPatients() {
    try {
        const search = document.getElementById('patientSearch').value;
        const url = search ? `${API_BASE}/patients?search=${encodeURIComponent(search)}` : `${API_BASE}/patients`;
        
        const response = await fetch(url);
        if (response.ok) {
            const data = await response.json();
            displayPatients(data.patients);
        }
    } catch (error) {
        console.error('加载患者列表失败:', error);
    }
}

// 显示患者列表
function displayPatients(patients) {
    const tbody = document.getElementById('patientsTable');
    const html = patients.map(patient => `
        <tr>
            <td>${patient.id}</td>
            <td>${patient.name}</td>
            <td>${patient.gender}</td>
            <td>${patient.age}</td>
            <td>${patient.phone || '-'}</td>
            <td>
                <button class="btn btn-sm btn-outline-primary" onclick="editPatient(${patient.id})">
                    <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger" onclick="deletePatient(${patient.id})">
                    <i class="bi bi-trash"></i>
                </button>
            </td>
        </tr>
    `).join('');
    
    tbody.innerHTML = html;
}

// 加载药品列表
async function loadMedicines() {
    try {
        const search = document.getElementById('medicineSearch').value;
        const category = document.getElementById('medicineCategory').value;
        
        let url = `${API_BASE}/medicines`;
        const params = new URLSearchParams();
        if (search) params.append('search', search);
        if (category) params.append('category', category);
        if (params.toString()) url += '?' + params.toString();
        
        const response = await fetch(url);
        if (response.ok) {
            const data = await response.json();
            displayMedicines(data.medicines);
        }
        
        // 加载药品分类
        loadMedicineCategories();
    } catch (error) {
        console.error('加载药品列表失败:', error);
    }
}

// 显示药品列表
function displayMedicines(medicines) {
    const tbody = document.getElementById('medicinesTable');
    const html = medicines.map(medicine => `
        <tr>
            <td>${medicine.id}</td>
            <td>${medicine.name}</td>
            <td>${medicine.specification}</td>
            <td>${medicine.category || '-'}</td>
            <td>
                <span class="${medicine.stock <= medicine.min_stock ? 'text-danger' : ''}">${medicine.stock}</span>
            </td>
            <td>¥${medicine.price}</td>
            <td>
                <button class="btn btn-sm btn-outline-primary" onclick="editMedicine(${medicine.id})">
                    <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-outline-warning" onclick="updateStock(${medicine.id})">
                    <i class="bi bi-box"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteMedicine(${medicine.id})">
                    <i class="bi bi-trash"></i>
                </button>
            </td>
        </tr>
    `).join('');
    
    tbody.innerHTML = html;
}

// 加载药品分类
async function loadMedicineCategories() {
    try {
        const response = await fetch(`${API_BASE}/medicines/categories`);
        if (response.ok) {
            const data = await response.json();
            const select = document.getElementById('medicineCategory');
            const currentValue = select.value;
            
            select.innerHTML = '<option value="">所有分类</option>';
            data.categories.forEach(category => {
                const option = document.createElement('option');
                option.value = category;
                option.textContent = category;
                select.appendChild(option);
            });
            
            select.value = currentValue;
        }
    } catch (error) {
        console.error('加载药品分类失败:', error);
    }
}

// 加载处方列表
async function loadPrescriptions() {
    try {
        const search = document.getElementById('prescriptionSearch').value;
        const status = document.getElementById('prescriptionStatus').value;
        
        let url = `${API_BASE}/prescriptions`;
        const params = new URLSearchParams();
        if (search) params.append('search', search);
        if (status) params.append('status', status);
        if (params.toString()) url += '?' + params.toString();
        
        const response = await fetch(url);
        if (response.ok) {
            const data = await response.json();
            displayPrescriptions(data.prescriptions);
        }
    } catch (error) {
        console.error('加载处方列表失败:', error);
    }
}

// 显示处方列表
function displayPrescriptions(prescriptions) {
    const tbody = document.getElementById('prescriptionsTable');
    const html = prescriptions.map(prescription => `
        <tr>
            <td>${prescription.id}</td>
            <td>${prescription.patient.name}</td>
            <td>${prescription.doctor.name}</td>
            <td>${prescription.diagnosis || '-'}</td>
            <td>${prescription.doctor_advice ? (prescription.doctor_advice.length > 20 ? prescription.doctor_advice.substring(0, 20) + '...' : prescription.doctor_advice) : '-'}</td>
            <td>¥${prescription.total_amount}</td>
            <td>
                <span class="badge bg-${getStatusBadgeColor(prescription.status)}">
                    ${getStatusText(prescription.status)}
                </span>
            </td>
            <td>${new Date(prescription.created_at).toLocaleDateString()}</td>
            <td>
                <button class="btn btn-sm btn-outline-primary" onclick="viewPrescription(${prescription.id})">
                    <i class="bi bi-eye"></i>
                </button>
                <button class="btn btn-sm btn-outline-success" onclick="printPrescription(${prescription.id})">
                    <i class="bi bi-printer"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger" onclick="deletePrescription(${prescription.id})">
                    <i class="bi bi-trash"></i>
                </button>
            </td>
        </tr>
    `).join('');
    
    tbody.innerHTML = html;
}

// 加载预约列表
async function loadAppointments() {
    try {
        const search = document.getElementById('appointmentSearch').value;
        const date = document.getElementById('appointmentDate').value;
        
        let url = `${API_BASE}/appointments`;
        const params = new URLSearchParams();
        if (search) params.append('search', search);
        if (date) params.append('date', date);
        if (params.toString()) url += '?' + params.toString();
        
        const response = await fetch(url);
        if (response.ok) {
            const data = await response.json();
            displayAppointments(data.appointments);
        }
    } catch (error) {
        console.error('加载预约列表失败:', error);
    }
}

// 显示预约列表
function displayAppointments(appointments) {
    const tbody = document.getElementById('appointmentsTable');
    const html = appointments.map(appointment => `
        <tr>
            <td>${appointment.id}</td>
            <td>${appointment.patient.name}</td>
            <td>${appointment.doctor.name}</td>
            <td>${new Date(appointment.appointment_time).toLocaleString()}</td>
            <td>${appointment.duration}分钟</td>
            <td>
                <span class="badge bg-${getAppointmentStatusBadgeColor(appointment.status)}">
                    ${getAppointmentStatusText(appointment.status)}
                </span>
            </td>
            <td>
                <button class="btn btn-sm btn-outline-primary" onclick="editAppointment(${appointment.id})">
                    <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-outline-success" onclick="printAppointment(${appointment.id})">
                    <i class="bi bi-printer"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteAppointment(${appointment.id})">
                    <i class="bi bi-trash"></i>
                </button>
            </td>
        </tr>
    `).join('');
    
    tbody.innerHTML = html;
}

// 加载统计报表
function loadReports() {
    // 这里可以添加图表显示逻辑
    console.log('加载统计报表');
}

// 工具函数
function getStatusBadgeColor(status) {
    const colors = {
        draft: 'secondary',
        completed: 'success',
        printed: 'info'
    };
    return colors[status] || 'secondary';
}

function getStatusText(status) {
    const texts = {
        draft: '草稿',
        completed: '已完成',
        printed: '已打印'
    };
    return texts[status] || status;
}

function getAppointmentStatusBadgeColor(status) {
    const colors = {
        scheduled: 'primary',
        completed: 'success',
        cancelled: 'danger',
        no_show: 'warning'
    };
    return colors[status] || 'secondary';
}

function getAppointmentStatusText(status) {
    const texts = {
        scheduled: '已预约',
        completed: '已完成',
        cancelled: '已取消',
        no_show: '未到'
    };
    return texts[status] || status;
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// 模态框相关函数
function showPatientModal(patientId = null) {
    const modal = new bootstrap.Modal(document.getElementById('patientModal'));
    const title = document.getElementById('patientModalTitle');
    const form = document.getElementById('patientForm');
    
    if (patientId) {
        title.textContent = '编辑患者';
        // 加载患者数据
        loadPatientData(patientId);
    } else {
        title.textContent = '添加患者';
        form.reset();
        document.getElementById('patientId').value = '';
    }
    
    modal.show();
}

function showMedicineModal(medicineId = null) {
    const modal = new bootstrap.Modal(document.getElementById('medicineModal'));
    const title = document.getElementById('medicineModalTitle');
    const form = document.getElementById('medicineForm');
    
    if (medicineId) {
        title.textContent = '编辑药品';
        // 加载药品数据
        loadMedicineData(medicineId);
    } else {
        title.textContent = '添加药品';
        form.reset();
        document.getElementById('medicineId').value = '';
    }
    
    modal.show();
}

function showPrescriptionModal(prescriptionId = null) {
    const modal = new bootstrap.Modal(document.getElementById('prescriptionModal'));
    const title = document.getElementById('prescriptionModalTitle');
    const form = document.getElementById('prescriptionForm');
    
    if (prescriptionId) {
        title.textContent = '编辑处方';
        // 加载处方数据
        loadPrescriptionData(prescriptionId);
    } else {
        title.textContent = '开处方';
        form.reset();
        document.getElementById('prescriptionId').value = '';
        // 重置处方明细
        document.getElementById('prescriptionItems').innerHTML = `
            <div class="row prescription-item">
                <div class="col-md-3 mb-2">
                    <input type="text" class="form-control" placeholder="药品名称" name="medicineName">
                </div>
                <div class="col-md-2 mb-2">
                    <input type="text" class="form-control" placeholder="规格" name="specification">
                </div>
                <div class="col-md-1 mb-2">
                    <input type="text" class="form-control" placeholder="用法" name="usage">
                </div>
                <div class="col-md-1 mb-2">
                    <input type="text" class="form-control" placeholder="频次" name="frequency">
                </div>
                <div class="col-md-1 mb-2">
                    <input type="number" class="form-control" placeholder="天数" name="days" value="1">
                </div>
                <div class="col-md-1 mb-2">
                    <input type="number" class="form-control" placeholder="数量" name="quantity" value="1">
                </div>
                <div class="col-md-1 mb-2">
                    <input type="number" step="0.01" class="form-control" placeholder="单价" name="unitPrice">
                </div>
                <div class="col-md-1 mb-2">
                    <button type="button" class="btn btn-danger btn-sm" onclick="removePrescriptionItem(this)">
                        <i class="bi bi-trash"></i>
                    </button>
                </div>
            </div>
        `;
    }
    
    // 加载患者和医生列表
    loadPatientsForSelect();
    loadDoctorsForSelect();
    
    modal.show();
}

function showAppointmentModal(appointmentId = null) {
    const modal = new bootstrap.Modal(document.getElementById('appointmentModal'));
    const title = document.getElementById('appointmentModalTitle');
    const form = document.getElementById('appointmentForm');
    
    if (appointmentId) {
        title.textContent = '编辑预约';
        // 加载预约数据
        loadAppointmentData(appointmentId);
    } else {
        title.textContent = '添加预约';
        form.reset();
        document.getElementById('appointmentId').value = '';
    }
    
    // 加载患者和医生列表
    loadPatientsForSelect();
    loadDoctorsForSelect();
    
    modal.show();
}

// 保存函数
async function savePatient() {
    const formData = {
        name: document.getElementById('patientName').value,
        gender: document.getElementById('patientGender').value,
        age: parseInt(document.getElementById('patientAge').value),
        phone: document.getElementById('patientPhone').value,
        id_card: document.getElementById('patientIdCard').value,
        address: document.getElementById('patientAddress').value,
        medical_history: document.getElementById('patientMedicalHistory').value
    };
    
    const patientId = document.getElementById('patientId').value;
    const url = patientId ? `${API_BASE}/patients/${patientId}` : `${API_BASE}/patients`;
    const method = patientId ? 'PUT' : 'POST';
    
    try {
        const response = await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });
        
        if (response.ok) {
            bootstrap.Modal.getInstance(document.getElementById('patientModal')).hide();
            loadPatients();
            alert(patientId ? '患者信息更新成功' : '患者添加成功');
        } else {
            const error = await response.json();
            alert(error.error || '操作失败');
        }
    } catch (error) {
        console.error('保存患者失败:', error);
        alert('操作失败，请检查网络连接');
    }
}

async function saveMedicine() {
    const formData = {
        name: document.getElementById('medicineName').value,
        specification: document.getElementById('medicineSpecification').value,
        unit: document.getElementById('medicineUnit').value,
        price: parseFloat(document.getElementById('medicinePrice').value),
        stock: parseInt(document.getElementById('medicineStock').value),
        category: document.getElementById('medicineCategory').value,
        manufacturer: document.getElementById('medicineManufacturer').value,
        min_stock: parseInt(document.getElementById('medicineMinStock').value)
    };
    
    const medicineId = document.getElementById('medicineId').value;
    const url = medicineId ? `${API_BASE}/medicines/${medicineId}` : `${API_BASE}/medicines`;
    const method = medicineId ? 'PUT' : 'POST';
    
    try {
        const response = await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });
        
        if (response.ok) {
            bootstrap.Modal.getInstance(document.getElementById('medicineModal')).hide();
            loadMedicines();
            alert(medicineId ? '药品信息更新成功' : '药品添加成功');
        } else {
            const error = await response.json();
            alert(error.error || '操作失败');
        }
    } catch (error) {
        console.error('保存药品失败:', error);
        alert('操作失败，请检查网络连接');
    }
}

async function savePrescription() {
    // 收集处方明细
    const items = [];
    document.querySelectorAll('.prescription-item').forEach(item => {
        const medicineName = item.querySelector('[name="medicineName"]').value;
        if (medicineName) {
            items.push({
                medicine_name: medicineName,
                specification: item.querySelector('[name="specification"]').value,
                usage: item.querySelector('[name="usage"]').value,
                frequency: item.querySelector('[name="frequency"]').value,
                days: parseInt(item.querySelector('[name="days"]').value),
                quantity: parseInt(item.querySelector('[name="quantity"]').value),
                unit_price: parseFloat(item.querySelector('[name="unitPrice"]').value) || 0
            });
        }
    });
    
    const formData = {
        patient_id: parseInt(document.getElementById('prescriptionPatient').value),
        diagnosis: document.getElementById('prescriptionDiagnosis').value,
        doctor_advice: document.getElementById('prescriptionDoctorAdvice').value,
        notes: document.getElementById('prescriptionNotes').value,
        items: items
    };
    
    const prescriptionId = document.getElementById('prescriptionId').value;
    const url = prescriptionId ? `${API_BASE}/prescriptions/${prescriptionId}` : `${API_BASE}/prescriptions`;
    const method = prescriptionId ? 'PUT' : 'POST';
    
    try {
        const response = await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });
        
        if (response.ok) {
            bootstrap.Modal.getInstance(document.getElementById('prescriptionModal')).hide();
            loadPrescriptions();
            alert(prescriptionId ? '处方更新成功' : '处方创建成功');
        } else {
            const error = await response.json();
            alert(error.error || '操作失败');
        }
    } catch (error) {
        console.error('保存处方失败:', error);
        alert('操作失败，请检查网络连接');
    }
}

async function saveAppointment() {
    const formData = {
        patient_id: parseInt(document.getElementById('appointmentPatient').value),
        doctor_id: parseInt(document.getElementById('appointmentDoctor').value),
        appointment_time: document.getElementById('appointmentTime').value,
        duration: parseInt(document.getElementById('appointmentDuration').value),
        notes: document.getElementById('appointmentNotes').value
    };
    
    const appointmentId = document.getElementById('appointmentId').value;
    const url = appointmentId ? `${API_BASE}/appointments/${appointmentId}` : `${API_BASE}/appointments`;
    const method = appointmentId ? 'PUT' : 'POST';
    
    try {
        const response = await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });
        
        if (response.ok) {
            bootstrap.Modal.getInstance(document.getElementById('appointmentModal')).hide();
            loadAppointments();
            alert(appointmentId ? '预约更新成功' : '预约创建成功');
        } else {
            const error = await response.json();
            alert(error.error || '操作失败');
        }
    } catch (error) {
        console.error('保存预约失败:', error);
        alert('操作失败，请检查网络连接');
    }
}

// 处方明细相关函数
function addPrescriptionItem() {
    const container = document.getElementById('prescriptionItems');
    const newItem = document.createElement('div');
    newItem.className = 'row prescription-item';
    newItem.innerHTML = `
        <div class="col-md-3 mb-2">
            <input type="text" class="form-control" placeholder="药品名称" name="medicineName">
        </div>
        <div class="col-md-2 mb-2">
            <input type="text" class="form-control" placeholder="规格" name="specification">
        </div>
        <div class="col-md-1 mb-2">
            <input type="text" class="form-control" placeholder="用法" name="usage">
        </div>
        <div class="col-md-1 mb-2">
            <input type="text" class="form-control" placeholder="频次" name="frequency">
        </div>
        <div class="col-md-1 mb-2">
            <input type="number" class="form-control" placeholder="天数" name="days" value="1">
        </div>
        <div class="col-md-1 mb-2">
            <input type="number" class="form-control" placeholder="数量" name="quantity" value="1">
        </div>
        <div class="col-md-1 mb-2">
            <input type="number" step="0.01" class="form-control" placeholder="单价" name="unitPrice">
        </div>
        <div class="col-md-1 mb-2">
            <button type="button" class="btn btn-danger btn-sm" onclick="removePrescriptionItem(this)">
                <i class="bi bi-trash"></i>
            </button>
        </div>
    `;
    container.appendChild(newItem);
}

function removePrescriptionItem(button) {
    const item = button.closest('.prescription-item');
    if (document.querySelectorAll('.prescription-item').length > 1) {
        item.remove();
    }
}

// 加载选择框数据
async function loadPatientsForSelect() {
    try {
        const response = await fetch(`${API_BASE}/patients`);
        if (response.ok) {
            const data = await response.json();
            const select = document.getElementById('prescriptionPatient') || document.getElementById('appointmentPatient');
            if (select) {
                select.innerHTML = '<option value="">请选择患者</option>';
                data.patients.forEach(patient => {
                    const option = document.createElement('option');
                    option.value = patient.id;
                    option.textContent = patient.name;
                    select.appendChild(option);
                });
            }
        }
    } catch (error) {
        console.error('加载患者列表失败:', error);
    }
}

async function loadDoctorsForSelect() {
    // 这里应该加载医生列表，暂时使用当前用户
    const select = document.getElementById('appointmentDoctor');
    if (select && currentUser) {
        select.innerHTML = '<option value="">请选择医生</option>';
        const option = document.createElement('option');
        option.value = currentUser.id;
        option.textContent = currentUser.name;
        select.appendChild(option);
    }
}

// 编辑和删除函数
function editPatient(id) {
    showPatientModal(id);
}

async function deletePatient(id) {
    if (!confirm('确定要删除这个患者吗？')) return;
    
    try {
        const response = await fetch(`${API_BASE}/patients/${id}`, { method: 'DELETE' });
        if (response.ok) {
            loadPatients();
            alert('患者删除成功');
        } else {
            const error = await response.json();
            alert(error.error || '删除失败');
        }
    } catch (error) {
        console.error('删除患者失败:', error);
        alert('删除失败，请检查网络连接');
    }
}

function editMedicine(id) {
    showMedicineModal(id);
}

async function updateStock(id) {
    const stock = prompt('请输入新的库存数量:');
    if (stock === null) return;
    
    try {
        const response = await fetch(`${API_BASE}/medicines/${id}/stock`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ stock: parseInt(stock) })
        });
        
        if (response.ok) {
            loadMedicines();
            alert('库存更新成功');
        } else {
            const error = await response.json();
            alert(error.error || '更新失败');
        }
    } catch (error) {
        console.error('更新库存失败:', error);
        alert('更新失败，请检查网络连接');
    }
}

async function deleteMedicine(id) {
    if (!confirm('确定要删除这个药品吗？')) return;
    
    try {
        const response = await fetch(`${API_BASE}/medicines/${id}`, { method: 'DELETE' });
        if (response.ok) {
            loadMedicines();
            alert('药品删除成功');
        } else {
            const error = await response.json();
            alert(error.error || '删除失败');
        }
    } catch (error) {
        console.error('删除药品失败:', error);
        alert('删除失败，请检查网络连接');
    }
}

function viewPrescription(id) {
    // 这里可以实现查看处方详情的功能
    alert('查看处方详情功能待实现');
}

async function printPrescription(id) {
    try {
        window.open(`${API_BASE}/print/prescription/${id}`, '_blank');
    } catch (error) {
        console.error('打印处方失败:', error);
        alert('打印失败');
    }
}

async function deletePrescription(id) {
    if (!confirm('确定要删除这个处方吗？')) return;
    
    try {
        const response = await fetch(`${API_BASE}/prescriptions/${id}`, { method: 'DELETE' });
        if (response.ok) {
            loadPrescriptions();
            alert('处方删除成功');
        } else {
            const error = await response.json();
            alert(error.error || '删除失败');
        }
    } catch (error) {
        console.error('删除处方失败:', error);
        alert('删除失败，请检查网络连接');
    }
}

function editAppointment(id) {
    showAppointmentModal(id);
}

async function printAppointment(id) {
    try {
        window.open(`${API_BASE}/print/appointment/${id}`, '_blank');
    } catch (error) {
        console.error('打印预约失败:', error);
        alert('打印失败');
    }
}

async function deleteAppointment(id) {
    if (!confirm('确定要删除这个预约吗？')) return;
    
    try {
        const response = await fetch(`${API_BASE}/appointments/${id}`, { method: 'DELETE' });
        if (response.ok) {
            loadAppointments();
            alert('预约删除成功');
        } else {
            const error = await response.json();
            alert(error.error || '删除失败');
        }
    } catch (error) {
        console.error('删除预约失败:', error);
        alert('删除失败，请检查网络连接');
    }
}

// 快速查找或创建患者
async function findOrCreatePatient() {
    const patientName = document.getElementById('prescriptionPatientName').value.trim();
    if (!patientName) {
        alert('请输入患者姓名');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/patients/find-or-create`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: patientName,
                gender: '',
                age: 0,
                phone: ''
            }),
        });

        if (response.ok) {
            const data = await response.json();
            const patient = data.patient;
            
            // 更新患者选择框
            const select = document.getElementById('prescriptionPatient');
            select.innerHTML = '<option value="">请选择患者</option>';
            
            const option = document.createElement('option');
            option.value = patient.id;
            option.textContent = `${patient.name} (${patient.gender || '未知'}, ${patient.age || '未知'}岁)`;
            select.appendChild(option);
            select.value = patient.id;
            
            // 清空输入框
            document.getElementById('prescriptionPatientName').value = '';
            
            // 显示提示信息
            const message = data.created ? '患者创建成功' : '找到现有患者';
            showToast(message, 'success');
        } else {
            const error = await response.json();
            alert(error.error || '查找患者失败');
        }
    } catch (error) {
        console.error('查找患者失败:', error);
        alert('查找患者失败，请检查网络连接');
    }
}

// 显示提示信息
function showToast(message, type = 'info') {
    // 创建提示元素
    const toast = document.createElement('div');
    toast.className = `alert alert-${type} alert-dismissible fade show position-fixed`;
    toast.style.cssText = 'top: 20px; right: 20px; z-index: 9999; min-width: 300px;';
    toast.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    document.body.appendChild(toast);
    
    // 3秒后自动移除
    setTimeout(() => {
        if (toast.parentNode) {
            toast.parentNode.removeChild(toast);
        }
    }, 3000);
}

// 加载处方数据用于编辑
async function loadPrescriptionData(prescriptionId) {
    try {
        const response = await fetch(`${API_BASE}/prescriptions/${prescriptionId}`);
        if (response.ok) {
            const data = await response.json();
            const prescription = data.prescription;
            
            // 填充表单数据
            document.getElementById('prescriptionId').value = prescription.id;
            document.getElementById('prescriptionDiagnosis').value = prescription.diagnosis || '';
            document.getElementById('prescriptionDoctorAdvice').value = prescription.doctor_advice || '';
            document.getElementById('prescriptionNotes').value = prescription.notes || '';
            
            // 设置患者选择
            if (prescription.patient) {
                const select = document.getElementById('prescriptionPatient');
                select.innerHTML = '<option value="">请选择患者</option>';
                const option = document.createElement('option');
                option.value = prescription.patient.id;
                option.textContent = prescription.patient.name;
                select.appendChild(option);
                select.value = prescription.patient.id;
            }
            
            // 加载处方明细
            const itemsContainer = document.getElementById('prescriptionItems');
            itemsContainer.innerHTML = '';
            
            if (prescription.items && prescription.items.length > 0) {
                prescription.items.forEach(item => {
                    const itemDiv = document.createElement('div');
                    itemDiv.className = 'row prescription-item';
                    itemDiv.innerHTML = `
                        <div class="col-md-3 mb-2">
                            <input type="text" class="form-control" placeholder="药品名称" name="medicineName" value="${item.medicine_name || ''}">
                        </div>
                        <div class="col-md-2 mb-2">
                            <input type="text" class="form-control" placeholder="规格" name="specification" value="${item.specification || ''}">
                        </div>
                        <div class="col-md-1 mb-2">
                            <input type="text" class="form-control" placeholder="用法" name="usage" value="${item.usage || ''}">
                        </div>
                        <div class="col-md-1 mb-2">
                            <input type="text" class="form-control" placeholder="频次" name="frequency" value="${item.frequency || ''}">
                        </div>
                        <div class="col-md-1 mb-2">
                            <input type="number" class="form-control" placeholder="天数" name="days" value="${item.days || 1}">
                        </div>
                        <div class="col-md-1 mb-2">
                            <input type="number" class="form-control" placeholder="数量" name="quantity" value="${item.quantity || 1}">
                        </div>
                        <div class="col-md-1 mb-2">
                            <input type="number" step="0.01" class="form-control" placeholder="单价" name="unitPrice" value="${item.unit_price || 0}">
                        </div>
                        <div class="col-md-1 mb-2">
                            <button type="button" class="btn btn-danger btn-sm" onclick="removePrescriptionItem(this)">
                                <i class="bi bi-trash"></i>
                            </button>
                        </div>
                    `;
                    itemsContainer.appendChild(itemDiv);
                });
            } else {
                // 如果没有明细，添加一个空的明细行
                addPrescriptionItem();
            }
        } else {
            alert('加载处方数据失败');
        }
    } catch (error) {
        console.error('加载处方数据失败:', error);
        alert('加载处方数据失败，请检查网络连接');
    }
}

// 加载患者数据用于编辑
async function loadPatientData(patientId) {
    try {
        const response = await fetch(`${API_BASE}/patients/${patientId}`);
        if (response.ok) {
            const data = await response.json();
            const patient = data.patient;
            
            // 填充表单数据
            document.getElementById('patientId').value = patient.id;
            document.getElementById('patientName').value = patient.name;
            document.getElementById('patientGender').value = patient.gender;
            document.getElementById('patientAge').value = patient.age;
            document.getElementById('patientPhone').value = patient.phone || '';
            document.getElementById('patientIdCard').value = patient.id_card || '';
            document.getElementById('patientAddress').value = patient.address || '';
            document.getElementById('patientMedicalHistory').value = patient.medical_history || '';
        } else {
            alert('加载患者数据失败');
        }
    } catch (error) {
        console.error('加载患者数据失败:', error);
        alert('加载患者数据失败，请检查网络连接');
    }
}

// 加载药品数据用于编辑
async function loadMedicineData(medicineId) {
    try {
        const response = await fetch(`${API_BASE}/medicines/${medicineId}`);
        if (response.ok) {
            const data = await response.json();
            const medicine = data.medicine;
            
            // 填充表单数据
            document.getElementById('medicineId').value = medicine.id;
            document.getElementById('medicineName').value = medicine.name;
            document.getElementById('medicineSpecification').value = medicine.specification;
            document.getElementById('medicineUnit').value = medicine.unit;
            document.getElementById('medicinePrice').value = medicine.price;
            document.getElementById('medicineStock').value = medicine.stock;
            document.getElementById('medicineCategory').value = medicine.category || '';
            document.getElementById('medicineManufacturer').value = medicine.manufacturer || '';
            document.getElementById('medicineMinStock').value = medicine.min_stock || 10;
        } else {
            alert('加载药品数据失败');
        }
    } catch (error) {
        console.error('加载药品数据失败:', error);
        alert('加载药品数据失败，请检查网络连接');
    }
}

// 快速添加药品
async function quickAddMedicine() {
    const medicineName = document.getElementById('quickMedicineName').value.trim();
    const specification = document.getElementById('quickMedicineSpec').value.trim();
    const price = parseFloat(document.getElementById('quickMedicinePrice').value) || 0;
    
    if (!medicineName || !specification) {
        alert('请填写药品名称和规格');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/medicines`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: medicineName,
                specification: specification,
                unit: '盒',
                price: price,
                stock: 0,
                category: '其他',
                manufacturer: '',
                min_stock: 10
            }),
        });

        if (response.ok) {
            const data = await response.json();
            
            // 更新当前药品输入框
            const currentMedicineInput = document.querySelector('.prescription-item:last-child [name="medicineName"]');
            if (currentMedicineInput) {
                currentMedicineInput.value = medicineName;
            }
            
            // 清空快速添加表单
            document.getElementById('quickMedicineName').value = '';
            document.getElementById('quickMedicineSpec').value = '';
            document.getElementById('quickMedicinePrice').value = '';
            
            // 隐藏快速添加模态框
            bootstrap.Modal.getInstance(document.getElementById('quickMedicineModal')).hide();
            
            showToast('药品添加成功', 'success');
        } else {
            const error = await response.json();
            alert(error.error || '添加药品失败');
        }
    } catch (error) {
        console.error('添加药品失败:', error);
        alert('添加药品失败，请检查网络连接');
    }
}

// 显示快速添加药品模态框
function showQuickMedicineModal() {
    const modal = new bootstrap.Modal(document.getElementById('quickMedicineModal'));
    const form = document.getElementById('quickMedicineForm');
    form.reset();
    modal.show();
} 
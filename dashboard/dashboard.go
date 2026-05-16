package dashboard

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/AndresCarvajalx/notiflow/logger"
)

type Notification struct {
	ID          int
	Phone       string
	Name        string
	Description string
	Value       float64
	DaysOverdue int
	Message     string
	Status      string
	ErrorDetail sql.NullString
	SentAt      string
}

type DashboardData struct {
	Notifications []Notification
	NotifJSON     template.JS
	TodayCount    int
	WeekCount     int
	MonthCount    int
	ErrorCount    int
}

var db *sql.DB

func Init(database *sql.DB) {
	db = database
}

func StartServer(port string) {
	http.HandleFunc("/", DashboardHandler)
	logger.L.Info("Dashboard running on http://localhost:" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.L.Panic(err.Error())
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT
			id, telefono, nombre, descripcion,
			valor, dias_vencidos, mensaje_enviado,
			estado, error_detalle, fecha_envio
		FROM notificaciones
		ORDER BY fecha_envio DESC
		LIMIT 500
	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID, &n.Phone, &n.Name, &n.Description,
			&n.Value, &n.DaysOverdue, &n.Message,
			&n.Status, &n.ErrorDetail, &n.SentAt,
		)
		if err != nil {
			continue
		}
		notifications = append(notifications, n)
	}

	jsonBytes, err := json.Marshal(notifications)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := DashboardData{
		Notifications: notifications,
		NotifJSON:     template.JS(jsonBytes),
		TodayCount:    getCount("date(fecha_envio) = date('now')"),
		WeekCount:     getCount("fecha_envio >= datetime('now', '-7 day')"),
		MonthCount:    getCount("fecha_envio >= datetime('now', '-30 day')"),
		ErrorCount:    getCount("estado = 'error'"),
	}

	tmpl := template.Must(template.New("dashboard").Parse(htmlTemplate))
	tmpl.Execute(w, data)
}

func getCount(condition string) int {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM notificaciones WHERE ` + condition).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

var htmlTemplate = `<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Notiflow Dashboard</title>
<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
<link href="https://cdn.jsdelivr.net/npm/@tabler/icons-webfont@latest/tabler-icons.min.css" rel="stylesheet">
<style>
*{box-sizing:border-box;}
body{background:#0f172a;color:#e2e8f0;font-family:system-ui,sans-serif;margin:0;}
.sidebar{width:220px;background:#1e293b;height:100vh;position:fixed;top:0;left:0;padding:1.5rem 1rem;border-right:1px solid #334155;}
.sidebar-logo{font-size:16px;font-weight:600;color:#f1f5f9;display:flex;align-items:center;gap:8px;margin-bottom:2rem;}
.sidebar-nav a{display:flex;align-items:center;gap:10px;padding:8px 12px;border-radius:8px;color:#94a3b8;text-decoration:none;font-size:14px;margin-bottom:4px;transition:all 0.15s;}
.sidebar-nav a:hover,.sidebar-nav a.active{background:#334155;color:#f1f5f9;}
.main{margin-left:220px;padding:2rem;}
.page-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:1.5rem;}
.page-title{font-size:20px;font-weight:600;color:#f1f5f9;}
.btn-refresh{background:#1e293b;border:1px solid #334155;color:#94a3b8;padding:6px 14px;border-radius:8px;font-size:13px;cursor:pointer;display:flex;align-items:center;gap:6px;text-decoration:none;transition:all 0.15s;}
.btn-refresh:hover{background:#334155;color:#f1f5f9;}
.metrics{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-bottom:1.5rem;}
.metric-card{background:#1e293b;border:1px solid #334155;border-radius:12px;padding:1.2rem 1.25rem;}
.metric-label{font-size:12px;color:#64748b;margin-bottom:6px;display:flex;align-items:center;gap:6px;}
.metric-value{font-size:30px;font-weight:600;color:#f1f5f9;line-height:1;}
.metric-sub{font-size:11px;color:#475569;margin-top:4px;}
.metric-error .metric-value{color:#f87171;}
.controls-row{display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;gap:8px;flex-wrap:wrap;}
.filter-btns{display:flex;gap:6px;}
.filter-btn{font-size:13px;padding:5px 14px;border-radius:8px;border:1px solid #334155;background:transparent;color:#64748b;cursor:pointer;transition:all 0.15s;}
.filter-btn:hover{background:#1e293b;color:#f1f5f9;}
.filter-btn.active{background:#334155;color:#f1f5f9;border-color:#475569;font-weight:500;}
.controls-right{display:flex;gap:8px;align-items:center;}
.search-wrap{position:relative;}
.search-wrap i{position:absolute;left:10px;top:50%;transform:translateY(-50%);color:#475569;font-size:15px;pointer-events:none;}
.search-input{background:#1e293b;border:1px solid #334155;color:#e2e8f0;border-radius:8px;padding:6px 12px 6px 34px;font-size:13px;width:200px;}
.search-input:focus{outline:none;border-color:#475569;}
.filter-select{background:#1e293b;border:1px solid #334155;color:#94a3b8;border-radius:8px;padding:6px 10px;font-size:13px;cursor:pointer;}
.filter-select:focus{outline:none;}
.table-card{background:#1e293b;border:1px solid #334155;border-radius:12px;overflow:hidden;}
.table-header{padding:1rem 1.25rem;border-bottom:1px solid #334155;display:flex;align-items:center;justify-content:space-between;}
.table-title{font-size:14px;font-weight:500;color:#f1f5f9;}
.result-count{font-size:12px;color:#475569;}
table{width:100%;border-collapse:collapse;}
thead th{padding:10px 1.25rem;text-align:left;font-size:11px;font-weight:500;color:#475569;text-transform:uppercase;letter-spacing:0.05em;background:#162032;border-bottom:1px solid #334155;}
tbody tr{border-bottom:1px solid #1e3145;transition:background 0.1s;}
tbody tr:last-child{border-bottom:none;}
tbody tr:hover{background:#162032;}
td{padding:10px 1.25rem;font-size:13px;color:#e2e8f0;vertical-align:middle;}
td.muted{color:#64748b;}
.avatar{width:30px;height:30px;border-radius:50%;background:#1d4ed8;display:inline-flex;align-items:center;justify-content:center;font-size:10px;font-weight:600;color:#bfdbfe;flex-shrink:0;}
.name-cell{display:flex;align-items:center;gap:10px;}
.name-main{font-weight:500;color:#f1f5f9;font-size:13px;}
.name-sub{font-size:11px;color:#475569;}
.badge{display:inline-flex;align-items:center;gap:4px;font-size:11px;font-weight:500;padding:3px 8px;border-radius:99px;}
.badge-sent{background:#14532d;color:#86efac;}
.badge-error{background:#450a0a;color:#fca5a5;}
.badge-skip{background:#422006;color:#fcd34d;}
.dot{width:5px;height:5px;border-radius:50%;display:inline-block;}
.dot-sent{background:#86efac;}
.dot-error{background:#fca5a5;}
.dot-skip{background:#fcd34d;}
.days-pill{display:inline-block;font-size:11px;font-weight:500;padding:2px 8px;border-radius:99px;}
.days-high{background:#450a0a;color:#fca5a5;}
.days-mid{background:#422006;color:#fcd34d;}
.days-low{background:#1e293b;color:#64748b;border:1px solid #334155;}
.empty-state{padding:3rem;text-align:center;color:#475569;font-size:13px;}
.empty-state i{font-size:36px;display:block;margin-bottom:8px;}
.pagination-row{padding:0.75rem 1.25rem;display:flex;align-items:center;justify-content:space-between;border-top:1px solid #334155;}
.page-info{font-size:12px;color:#475569;}
.page-btns{display:flex;gap:4px;}
.page-btn{font-size:12px;padding:4px 12px;border-radius:8px;border:1px solid #334155;background:transparent;color:#94a3b8;cursor:pointer;}
.page-btn:hover{background:#334155;color:#f1f5f9;}
.page-btn:disabled{opacity:0.3;cursor:default;}
</style>
</head>
<body>

<div class="sidebar">
	<div class="sidebar-logo">
		<i class="ti ti-bell" style="font-size:18px;"></i>
		Notiflow
	</div>
	<nav class="sidebar-nav">
		<a href="#" class="active"><i class="ti ti-layout-dashboard"></i> Dashboard</a>
	</nav>
</div>

<div class="main">

	<div class="page-header">
		<span class="page-title">Dashboard</span>
		<button class="btn-refresh" onclick="location.reload()">
			<i class="ti ti-refresh"></i> Actualizar
		</button>
	</div>

	<div class="metrics">
		<div class="metric-card">
			<div class="metric-label"><i class="ti ti-calendar"></i> Hoy</div>
			<div class="metric-value">{{.TodayCount}}</div>
			<div class="metric-sub">notificaciones</div>
		</div>
		<div class="metric-card">
			<div class="metric-label"><i class="ti ti-calendar-week"></i> 7 días</div>
			<div class="metric-value">{{.WeekCount}}</div>
			<div class="metric-sub">notificaciones</div>
		</div>
		<div class="metric-card">
			<div class="metric-label"><i class="ti ti-calendar-month"></i> 30 días</div>
			<div class="metric-value">{{.MonthCount}}</div>
			<div class="metric-sub">notificaciones</div>
		</div>
		<div class="metric-card metric-error">
			<div class="metric-label"><i class="ti ti-alert-circle"></i> Errores</div>
			<div class="metric-value">{{.ErrorCount}}</div>
			<div class="metric-sub">totales</div>
		</div>
	</div>

	<div class="controls-row">
		<div class="filter-btns">
			<button class="filter-btn active" id="btn-today" onclick="setFilter('today')">Hoy</button>
			<button class="filter-btn" id="btn-7" onclick="setFilter('7')">7 días</button>
			<button class="filter-btn" id="btn-30" onclick="setFilter('30')">30 días</button>
		</div>
		<div class="controls-right">
			<select class="filter-select" id="status-filter" onchange="applyFilters()">
				<option value="">Todos los estados</option>
				<option value="enviado">Enviado</option>
				<option value="error">Error</option>
				<option value="omitido">Omitido</option>
			</select>
			<div class="search-wrap">
				<i class="ti ti-search"></i>
				<input class="search-input" type="text" id="search" placeholder="Buscar cliente..." oninput="applyFilters()">
			</div>
		</div>
	</div>

	<div class="table-card">
		<div class="table-header">
			<span class="table-title" id="table-title">Notificaciones de hoy</span>
			<span class="result-count" id="result-count"></span>
		</div>
		<table>
			<thead>
				<tr>
					<th>Cliente</th>
					<th>Teléfono</th>
					<th>Estado</th>
					<th>Días vencidos</th>
					<th>Fecha envío</th>
				</tr>
			</thead>
			<tbody id="table-body"></tbody>
		</table>
		<div id="empty-state" class="empty-state" style="display:none;">
			<i class="ti ti-inbox"></i>
			Sin notificaciones para este período
		</div>
		<div class="pagination-row">
			<span class="page-info" id="page-info"></span>
			<div class="page-btns">
				<button class="page-btn" id="btn-prev" onclick="changePage(-1)">← Anterior</button>
				<button class="page-btn" id="btn-next" onclick="changePage(1)">Siguiente →</button>
			</div>
		</div>
	</div>

</div>

<script>
const ALL_DATA = {{.NotifJSON}};
const PAGE_SIZE = 10;
let currentFilter = 'today';
let currentPage = 1;
let filtered = [];

function initials(name) {
	return (name || '').split(' ').slice(0,2).map(w => w[0] || '').join('').toUpperCase() || '?';
}

function statusBadge(s) {
	if (s === 'enviado') return '<span class="badge badge-sent"><span class="dot dot-sent"></span>Enviado</span>';
	if (s === 'error')   return '<span class="badge badge-error"><span class="dot dot-error"></span>Error</span>';
	return '<span class="badge badge-skip"><span class="dot dot-skip"></span>Omitido</span>';
}

function daysPill(d) {
	if (d >= 60) return '<span class="days-pill days-high">'+d+' días</span>';
	if (d >= 30) return '<span class="days-pill days-mid">'+d+' días</span>';
	return '<span class="days-pill days-low">'+d+' días</span>';
}

function inRange(item) {
	if (!item.SentAt) return false;
	const date = new Date(item.SentAt.replace(' ', 'T'));
	const now = new Date();
	const today = now.toISOString().slice(0,10);
	if (currentFilter === 'today') return item.SentAt.slice(0,10) === today;
	const days = currentFilter === '7' ? 7 : 30;
	const limit = new Date(now.getTime() - days * 86400000);
	return date >= limit;
}

function applyFilters() {
	const q = document.getElementById('search').value.toLowerCase();
	const st = document.getElementById('status-filter').value;
	filtered = ALL_DATA.filter(item => {
		if (!inRange(item)) return false;
		if (st && item.Status !== st) return false;
		if (q && !(item.Name||'').toLowerCase().includes(q) && !(item.Phone||'').includes(q)) return false;
		return true;
	});
	currentPage = 1;
	renderTable();
}

function setFilter(range) {
	currentFilter = range;
	document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
	document.getElementById('btn-'+range).classList.add('active');
	const titles = { today: 'Notificaciones de hoy', '7': 'Últimos 7 días', '30': 'Últimos 30 días' };
	document.getElementById('table-title').textContent = titles[range];
	applyFilters();
}

function changePage(dir) {
	const maxPage = Math.ceil(filtered.length / PAGE_SIZE);
	currentPage = Math.min(Math.max(1, currentPage + dir), maxPage);
	renderTable();
}

function renderTable() {
	const total = filtered.length;
	const start = (currentPage - 1) * PAGE_SIZE;
	const page = filtered.slice(start, start + PAGE_SIZE);
	const tbody = document.getElementById('table-body');
	const empty = document.getElementById('empty-state');
	const maxPage = Math.ceil(total / PAGE_SIZE) || 1;

	document.getElementById('result-count').textContent = total + ' resultado' + (total !== 1 ? 's' : '');
	document.getElementById('page-info').textContent = total > 0 ? (start+1)+'\u2013'+Math.min(start+PAGE_SIZE, total)+' de '+total : '';
	document.getElementById('btn-prev').disabled = currentPage <= 1;
	document.getElementById('btn-next').disabled = currentPage >= maxPage;

	if (page.length === 0) {
		tbody.innerHTML = '';
		empty.style.display = 'block';
		return;
	}
	empty.style.display = 'none';

	tbody.innerHTML = page.map(n => {
		const av = initials(n.Name);
		return '<tr>' +
			'<td><div class="name-cell"><div class="avatar">'+av+'</div><div><div class="name-main">'+(n.Name||'')+'</div><div class="name-sub">'+(n.Description||'')+'</div></div></div></td>' +
			'<td class="muted">'+(n.Phone||'')+'</td>' +
			'<td>'+statusBadge(n.Status)+'</td>' +
			'<td>'+daysPill(n.DaysOverdue)+'</td>' +
			'<td class="muted">'+(n.SentAt||'')+'</td>' +
		'</tr>';
	}).join('');
}

setFilter('today');
</script>
</body>
</html>
`

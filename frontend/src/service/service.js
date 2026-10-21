// 格式化字节数
function formatBytes(bytes, decimals = 2) {
    if (!bytes || bytes <= 0 || !Number.isFinite(bytes)) return '0 Bytes';

    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

// 格式化时间
function formatDate(date) {
    if (!date) return '—';
    const d = date instanceof Date ? date : new Date(date);
    if (isNaN(d.getTime())) return '—';
    return d.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// input: [{"date":"20251230","traffic":0},{"date":"20251229","traffic":0}]
// return formated by formatDate
// if date is today, return formated traffic by formatBytes; otherwise return 0;   return formated by formatDate
function getTrafficOfTodayFromArray(dates) {
    // today is in format of YYYYMMDD
    const today = new Date().getFullYear().toString() + 
                (new Date().getMonth() + 1).toString().padStart(2, '0') + 
                new Date().getDate().toString().padStart(2, '0'); 
    if (!Array.isArray(dates) || dates.length === 0) return 0;
    const todayTraffic = dates.find(date => date.date === today);
    if (todayTraffic) {
        return formatBytes(todayTraffic.traffic);
    } else {
        return 0;
    }
}

// input: [{"month":"202512","traffic":0},{"month":"202511","traffic":0},{"month":"202510","traffic":0}]
// return formated by formatDate
// if month is current month, return formated traffic by formatBytes; otherwise return 0;   return formated by formatDate
function getCurrentMonthTraffic(months) {
    const currentMonth = new Date().getFullYear().toString() + 
                (new Date().getMonth() + 1).toString().padStart(2, '0');
    if (!Array.isArray(months) || months.length === 0) return 0;
    const currentMonthTraffic = months.find(month => month.month === currentMonth);
    if (currentMonthTraffic) {
        return formatBytes(currentMonthTraffic.traffic);
    } else {
        return 0;
    }
}

// input: [{"year":"2025","traffic":0},{"year":"2024","traffic":0},{"year":"2023","traffic":0}]
// return formated by formatDate
// if year is current year, return formated traffic by formatBytes; otherwise return 0;   return formated by formatDate
function getCurrentYearTraffic(years) {
    const currentYear = new Date().getFullYear().toString();
    if (!Array.isArray(years) || years.length === 0) return 0;
    const currentYearTraffic = years.find(year => year.year === currentYear);
    if (currentYearTraffic) {
        return formatBytes(currentYearTraffic.traffic);
    } else {    
        return 0;
    }
}

export {
    formatBytes,
    formatDate,
    getTrafficOfTodayFromArray,
    getCurrentMonthTraffic,
    getCurrentYearTraffic
}
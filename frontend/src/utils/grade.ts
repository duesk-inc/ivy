export function gradeColor(grade: string): string {
  switch (grade) {
    case 'A': return '#2E7D32';
    case 'B': return '#1565C0';
    case 'C': return '#E65100';
    case 'D': return '#C62828';
    default: return '#757575';
  }
}

export function skillStatusColor(status: string): string {
  switch (status) {
    case 'met': return '#2E7D32';
    case 'partial': return '#E65100';
    case 'unmet': return '#C62828';
    default: return '#757575';
  }
}

export function skillStatusLabel(status: string): string {
  switch (status) {
    case 'met': return '充足';
    case 'partial': return '部分的';
    case 'unmet': return '未充足';
    default: return '不明';
  }
}

export function ngStatusColor(status: string): string {
  switch (status) {
    case 'ok': return '#2E7D32';
    case 'ng': return '#C62828';
    case 'warning': return '#E65100';
    case 'unknown': return '#757575';
    default: return '#757575';
  }
}

export function ngStatusLabel(status: string): string {
  switch (status) {
    case 'ok': return 'OK';
    case 'ng': return 'NG';
    case 'warning': return '要確認';
    case 'unknown': return '不明';
    default: return status;
  }
}

export function supplyChainColor(level: number): string {
  switch (level) {
    case 1: return '#2E7D32';
    case 2: return '#1565C0';
    case 3: return '#E65100';
    case 4: return '#C62828';
    default: return '#757575';
  }
}

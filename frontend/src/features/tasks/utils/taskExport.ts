import dayjs from 'dayjs';
import type { Task } from '@orbito/types';
import type { ColumnConfig } from '../types/taskViewTypes';

/**
 * Extracts cell content for a given column and task.
 */
export function extractTaskFieldValue(task: Task, columnId: string): string {
  switch (columnId) {
    case 'slug':
      return task.slug || `TASK-${task.taskNumber || task.id.slice(0, 4)}`;

    case 'title':
      return task.title || '';

    case 'description':
      return (task.description || '').replace(/\r?\n/g, ' ').trim();

    case 'type':
      return task.taskType || 'TASK';

    case 'status':
      return task.status?.name || 'TODO';

    case 'priority':
      return task.priority || 'MEDIUM';

    case 'assignees':
      if (task.assignees && task.assignees.length > 0) {
        return task.assignees
          .map((a) => {
            const u = a.user;
            if (!u) return 'Unassigned';
            return `${u.firstName || ''} ${u.lastName || ''}`.trim() || u.username || 'User';
          })
          .join(', ');
      }
      return 'Unassigned';

    case 'sprint':
      return task.sprint?.name || 'Backlog';

    case 'storyPoints':
      return task.storyPoints != null ? String(task.storyPoints) : '';

    case 'dueDate':
      return task.dueDate ? dayjs(task.dueDate).format('YYYY-MM-DD') : '';

    case 'startDate':
      return task.startDate ? dayjs(task.startDate).format('YYYY-MM-DD') : '';

    case 'timeTracking':
      return `${task.loggedHours || 0}h / ${task.estimatedHours || 0}h`;

    case 'recurrence':
      return task.isRecurring ? 'Recurring' : 'None';

    case 'createdAt':
      return task.createdAt ? dayjs(task.createdAt).format('YYYY-MM-DD HH:mm') : '';

    default:
      return '';
  }
}

/**
 * Exports tasks to RFC-4180 compliant CSV file with UTF-8 BOM.
 */
export function exportTasksToCsv(
  tasks: Task[],
  columns: ColumnConfig[],
  filename: string = 'tasks_export.csv'
): void {
  try {
    const exportColumns = columns.filter((col) => col.visible);
    if (exportColumns.length === 0) return;

    // Header row
    const headers = exportColumns.map((col) => `"${col.label.replace(/"/g, '""')}"`);

    // Data rows
    const rows = tasks.map((task) =>
      exportColumns
        .map((col) => {
          const val = extractTaskFieldValue(task, col.id);
          const escaped = val.replace(/"/g, '""');
          return `"${escaped}"`;
        })
        .join(',')
    );

    // Combine UTF-8 BOM + headers + rows
    const csvContent = '\uFEFF' + [headers.join(','), ...rows].join('\r\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', filename.endsWith('.csv') ? filename : `${filename}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch (error) {
    console.error('Failed to export tasks to CSV:', error);
  }
}

/**
 * Exports tasks to Microsoft Excel XML Spreadsheet format (.xls) with formatted styles.
 */
export function exportTasksToExcel(
  tasks: Task[],
  columns: ColumnConfig[],
  filename: string = 'tasks_export.xls'
): void {
  try {
    const exportColumns = columns.filter((col) => col.visible);
    if (exportColumns.length === 0) return;

    const escapeXml = (unsafe: string) =>
      unsafe
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&apos;');

    let xml = `<?xml version="1.0" encoding="UTF-8"?>
<?mso-application progid="Excel.Sheet"?>
<Workbook xmlns="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:o="urn:schemas-microsoft-com:office:office"
 xmlns:x="urn:schemas-microsoft-com:office:excel"
 xmlns:ss="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:html="http://www.w3.org/TR/REC-html40">
 <Styles>
  <Style ss:ID="Default" ss:Name="Normal">
   <Alignment ss:Vertical="Center"/>
   <Borders/>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="11" ss:Color="#000000"/>
  </Style>
  <Style ss:ID="HeaderStyle">
   <Alignment ss:Horizontal="Left" ss:Vertical="Center"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#D4D4D8"/>
   </Borders>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="11" ss:Color="#18181B" ss:Bold="1"/>
   <Interior ss:Color="#F4F4F5" ss:Pattern="Solid"/>
  </Style>
 </Styles>
 <Worksheet ss:Name="Tasks">
  <Table>`;

    // Define column widths
    exportColumns.forEach((col) => {
      xml += `\n   <Column ss:Width="${Math.max(col.minWidth || 100, 100)}"/>`;
    });

    // Header row
    xml += '\n   <Row ss:Height="24">';
    exportColumns.forEach((col) => {
      xml += `\n    <Cell ss:StyleID="HeaderStyle"><Data ss:Type="String">${escapeXml(col.label)}</Data></Cell>`;
    });
    xml += '\n   </Row>';

    // Data rows
    tasks.forEach((task) => {
      xml += '\n   <Row ss:Height="20">';
      exportColumns.forEach((col) => {
        const value = extractTaskFieldValue(task, col.id);
        const isNumber = col.id === 'storyPoints' && !isNaN(Number(value)) && value.trim() !== '';
        const dataType = isNumber ? 'Number' : 'String';
        xml += `\n    <Cell><Data ss:Type="${dataType}">${escapeXml(value)}</Data></Cell>`;
      });
      xml += '\n   </Row>';
    });

    xml += `\n  </Table>
 </Worksheet>
</Workbook>`;

    const blob = new Blob([xml], { type: 'application/vnd.ms-excel;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', filename.endsWith('.xls') ? filename : `${filename}.xls`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch (error) {
    console.error('Failed to export tasks to Excel:', error);
  }
}


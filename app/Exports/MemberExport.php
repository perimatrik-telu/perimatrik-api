<?php

namespace App\Exports;

use App\Models\Member;
use Maatwebsite\Excel\Concerns\ShouldAutoSize;
use Maatwebsite\Excel\Concerns\WithCustomCsvSettings;
use Maatwebsite\Excel\Concerns\WithEvents;
use Maatwebsite\Excel\Concerns\FromCollection;
use Maatwebsite\Excel\Concerns\WithHeadings;
use Maatwebsite\Excel\Concerns\WithMapping;
use PhpOffice\PhpSpreadsheet\Style\Alignment;
use PhpOffice\PhpSpreadsheet\Style\Border;
use PhpOffice\PhpSpreadsheet\Style\Fill;
use Maatwebsite\Excel\Events\AfterSheet;

class MemberExport implements FromCollection, WithHeadings, WithMapping, ShouldAutoSize, WithEvents, WithCustomCsvSettings
{
    public function __construct(
        protected string $format = 'xlsx'
    ) {
    }

    /**
    * @return \Illuminate\Support\Collection
    */
    public function collection()
    {
        return Member::with('angkatan')
            ->join('angkatans', 'members.angkatan_id', '=', 'angkatans.id')
            ->orderBy('angkatans.year', 'desc')
            ->select('members.*')
            ->get();
    }

    public function headings(): array
    {
        return [
            'No Anggota',
            'Nama Lengkap',
            'Nama Angkatan',
            'Tahun Angkatan',
            'Nama Lapangan',
            'Domisili',
            'Kontak'
        ];
    }

    public function map($member): array
    {
        return [
            $member->no_anggota,
            $member->nama_lengkap,
            $member->angkatan->nama_angkatan ?? '',
            $member->angkatan->year ?? '',
            $member->nama_lapangan,
            $member->domisili,
            $member->kontak
        ];
    }

    public function registerEvents(): array
    {
        return [
            AfterSheet::class => function (AfterSheet $event) {
                if ($this->format === 'csv') {
                    return;
                }

                $sheet = $event->sheet->getDelegate();
                $highestRow = $sheet->getHighestRow();
                $highestColumn = $sheet->getHighestColumn();
                $headerRange = "A1:{$highestColumn}1";
                $dataRange = "A1:{$highestColumn}{$highestRow}";

                $sheet->freezePane('A2');
                $sheet->setAutoFilter($dataRange);

                $sheet->getStyle($headerRange)->applyFromArray([
                    'font' => [
                        'bold' => true,
                        'color' => ['rgb' => 'FFFFFF'],
                        'size' => 12,
                    ],
                    'fill' => [
                        'fillType' => Fill::FILL_SOLID,
                        'startColor' => ['rgb' => '1F4E78'],
                    ],
                    'alignment' => [
                        'horizontal' => Alignment::HORIZONTAL_CENTER,
                        'vertical' => Alignment::VERTICAL_CENTER,
                    ],
                    'borders' => [
                        'allBorders' => [
                            'borderStyle' => Border::BORDER_THIN,
                            'color' => ['rgb' => 'D9E2F3'],
                        ],
                    ],
                ]);

                $sheet->getStyle($dataRange)->applyFromArray([
                    'alignment' => [
                        'vertical' => Alignment::VERTICAL_CENTER,
                        'wrapText' => true,
                    ],
                    'borders' => [
                        'allBorders' => [
                            'borderStyle' => Border::BORDER_THIN,
                            'color' => ['rgb' => 'E2E8F0'],
                        ],
                    ],
                ]);

                $sheet->getRowDimension(1)->setRowHeight(24);
                $sheet->getSheetView()->setZoomScale(90);
                $sheet->setTitle('Data Anggota');

                foreach (['A' => 18, 'B' => 30, 'C' => 24, 'D' => 16, 'E' => 24, 'F' => 22, 'G' => 20] as $column => $width) {
                    $sheet->getColumnDimension($column)->setWidth($width);
                }
            },
        ];
    }

    public function getCsvSettings(): array
    {
        return [
            'delimiter' => ',',
            'enclosure' => '"',
            'line_ending' => PHP_EOL,
            'use_bom' => true,
            'excel_compatibility' => true,
            'output_encoding' => 'UTF-8',
        ];
    }
}

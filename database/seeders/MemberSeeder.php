<?php

namespace Database\Seeders;

use App\Models\Angkatan;
use App\Models\Member;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\File;

class MemberSeeder extends Seeder
{
    /**
     * Run the database seeds.
     */
    public function run(): void
    {
        $csvFile = File::get(base_path('MASTERSHEET HUBUNGAN MASYARAKAT - DATABASE ANGGOTA.csv'));
        $data = array_map('str_getcsv', explode("\n", $csvFile));
        $header = array_shift($data);

        foreach ($data as $row) {
            if (count($header) !== count($row)) {
                continue;
            }
            $rowData = array_combine($header, $row);

            $angkatanName = $rowData['Angkatan Perimatrik'];
            $angkatan = Angkatan::firstOrCreate(
                ['nama_angkatan' => $angkatanName],
                ['year' => rand(1990, 2023)]
            );

            Member::updateOrCreate(
                ['no_anggota' => $rowData['No Anggota']],
                [
                    'nama_lengkap' => $rowData['Nama Lengkap'],
                    'nama_lapangan' => $rowData['Nama Lapangan'],
                    'domisili' => $rowData['Domisili'],
                    'kontak' => $rowData['Kontak'],
                    'avatar' => null,
                    'angkatan_id' => $angkatan->id,
                ]
            );
        }
    }
}

<?php
namespace App\Http\Controllers;

use App\Enums\ResponseMessage;
use App\Enums\StatusCode;
use App\Helpers\AttachmentHelper;
use App\Helpers\ResponseHelper;
use App\Models\Member;
use App\Exports\MemberExport;
use Maatwebsite\Excel\Facades\Excel;
use Maatwebsite\Excel\Excel as ExcelFormat;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Validator;
use Illuminate\Validation\Rule;

class MemberController extends Controller
{
    /**
     * Display a listing of the resource.
     */
    public function index(Request $request)
    {
        $members = Member::with('angkatan')
            ->join('angkatans', 'members.angkatan_id', '=', 'angkatans.id')
            ->orderBy('angkatans.year', 'desc')
            ->select('members.*') // Select all columns from the members table to avoid conflicts
            ->get();
        return ResponseHelper::responseSuccessAllData($members, StatusCode::OK);
    }

    /**
     * Export members to Excel (.xlsx)
     */
    public function export(Request $request)
    {
        $format = strtolower($request->query('format', 'xlsx'));

        if ($format === 'csv') {
            return Excel::download(
                new MemberExport($format),
                'database_anggota_perimatrik.csv',
                ExcelFormat::CSV
            );
        }

        return Excel::download(new MemberExport($format), 'database_anggota_perimatrik.xlsx');
    }

    /**
     * Store a newly created resource in storage.
     */
    public function store(Request $request)
    {
        $rules = [
            'no_anggota'    => 'required|string|max:255|unique:members',
            'nama_lengkap'  => 'required|string|max:255',
            'nama_lapangan' => 'nullable|string|max:255',
            'domisili'      => 'nullable|string|max:255',
            'kontak'        => 'nullable|string|max:255',
            'angkatan_id'   => 'required|uuid|exists:angkatans,id',
            'user_id'       => 'nullable|uuid|exists:users,id',
            'instagram'     => 'nullable|string|max:255',
            'facebook'      => 'nullable|string|max:255',
            'linkedin'      => 'nullable|string|max:255',
            'tiktok'        => 'nullable|string|max:255',
            'youtube'       => 'nullable|string|max:255',
        ];

        if ($request->hasFile('avatar')) {
            $rules['avatar'] = 'required|file|mimes:jpeg,png,jpg,gif,svg|max:2048';
        } else {
            $rules['avatar'] = 'nullable|string|url';
        }

        $validator = Validator::make($request->all(), $rules);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                ResponseMessage::FAILED_CREATE,
                $validator->errors()
            );
        }

        $fullImageUrl = null;
        if ($request->hasFile('avatar')) {
            $filename     = AttachmentHelper::saveImageMedia($request, 'avatar', config('path.MEMBER_PATH'));
            $fullImageUrl = config('app.url') . '/storage/' . ltrim(config('path.MEMBER_PATH'), '/') . '/' . $filename;
        } else {
            $fullImageUrl = $request->input('avatar');
        }

        $memberData = array_merge(
            $request->except('avatar'),
            ['avatar' => $fullImageUrl]
        );

        $member = Member::create($memberData);
        return ResponseHelper::responseSuccessSingleData($member, StatusCode::CREATED);
    }

    /**
     * Display the specified resource.
     */
    public function show(string $id)
    {
        $member = Member::with('angkatan')->find($id);

        if (! $member) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                ResponseMessage::NO_ITEM
            );
        }

        return ResponseHelper::responseSuccessSingleData($member, StatusCode::OK);
    }

    /**
     * Update the specified resource in storage.
     */
    public function update(Request $request, string $id)
    {
        $member = Member::find($id);

        if (! $member) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                ResponseMessage::NO_ITEM
            );
        }

        $rules = [
            'no_anggota'    => ['required', 'string', 'max:255', Rule::unique('members')->ignore($member->id)],
            'nama_lengkap'  => 'required|string|max:255',
            'nama_lapangan' => 'nullable|string|max:255',
            'domisili'      => 'nullable|string|max:255',
            'kontak'        => 'nullable|string|max:255',
            'angkatan_id'   => 'required|uuid|exists:angkatans,id',
            'user_id'       => 'nullable|uuid|exists:users,id',
            'avatar'        => 'nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048',
            'instagram'     => 'nullable|string|max:255',
            'facebook'      => 'nullable|string|max:255',
            'linkedin'      => 'nullable|string|max:255',
            'tiktok'        => 'nullable|string|max:255',
            'youtube'       => 'nullable|string|max:255',
        ];

        $dataToValidate = $request->except(['avatar', '_method']);
        if ($request->hasFile('avatar')) {
            $dataToValidate['avatar'] = $request->file('avatar');
        }

        $validator = Validator::make($dataToValidate, $rules);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                ResponseMessage::FAILED_UPDATE,
                $validator->errors()
            );
        }

        $memberData = $request->except(['avatar', '_method']);

        if ($request->hasFile('avatar')) {
            if ($member->avatar && str_contains($member->avatar, '/storage/')) {
                AttachmentHelper::deleteMedia(config('path.MEMBER_PATH'), basename(parse_url($member->avatar, PHP_URL_PATH)));
            }
            $filename             = AttachmentHelper::saveImageMedia($request, 'avatar', config('path.MEMBER_PATH'));
            $memberData['avatar'] = config('app.url') . '/storage/' . ltrim(config('path.MEMBER_PATH'), '/') . '/' . $filename;
        }

        $member->update($memberData);
        return ResponseHelper::responseSuccessSingleData($member, StatusCode::OK);
    }

    /**
     * Remove the specified resource from storage.
     */
    public function destroy(string $id)
    {
        $member = Member::find($id);

        if (! $member) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                'Member not found'
            );
        }

        if ($member->avatar && str_contains($member->avatar, '/storage/')) {
            AttachmentHelper::deleteMedia(config('path.MEMBER_PATH'), basename(parse_url($member->avatar, PHP_URL_PATH)));
        }

        $member->delete();

        return ResponseHelper::responseMessageRequest(
            StatusCode::OK,
            true,
            'Member deleted successfully'
        );
    }
}

<?php
namespace App\Http\Controllers;

use App\Enums\StatusCode;
use App\Helpers\ResponseHelper;
use App\Models\Angkatan;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Validator;

class AngkatanController extends Controller
{
    /**
     * Display a listing of the resource.
     */
    public function index()
    {
        $angkatans = Angkatan::all();
        return ResponseHelper::responseSuccessAllData($angkatans, StatusCode::OK);
    }

    /**
     * Store a newly created resource in storage.
     */
    public function store(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'nama_angkatan' => 'required|string|max:255|unique:angkatans',
            'year'          => 'required|string|max:4',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                'Validation Error',
                $validator->errors()
            );
        }

        $angkatan = Angkatan::create($request->all());
        return ResponseHelper::responseSuccessSingleData(
            $angkatan,
            StatusCode::CREATED
        );
    }

    /**
     * Display the specified resource.
     */
    public function show(string $id)
    {
        $angkatan = Angkatan::find($id);

        if (! $angkatan) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                'Angkatan not found'
            );
        }

        return ResponseHelper::responseSuccessSingleData(
            $angkatan,
            StatusCode::OK
        );
    }

    /**
     * Update the specified resource in storage.
     */
    public function update(Request $request, string $id)
    {
        $angkatan = Angkatan::find($id);

        if (! $angkatan) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                'Angkatan not found'
            );
        }

        $validator = Validator::make($request->all(), [
            'nama_angkatan' => 'required|string|max:255|unique:angkatans,nama_angkatan,' . $id,
            'year'          => 'required|string|max:4',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                'Validation Error',
                $validator->errors()
            );
        }

        $angkatan->update($request->all());
        return ResponseHelper::responseSuccessSingleData(
            $angkatan,
            StatusCode::OK
        );
    }

    /**
     * Remove the specified resource from storage.
     */
    public function destroy(string $id)
    {
        $angkatan = Angkatan::find($id);

        if (! $angkatan) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                'Angkatan not found'
            );
        }

        $angkatan->delete();
        return ResponseHelper::responseMessageRequest(
            StatusCode::OK,
            true,
            'Angkatan deleted successfully'
        );
    }
}

<?php
namespace App\Helpers;

use App\Enums\ErrorCode;
use App\Enums\ResponseMessage;
use App\Enums\StatusCode;
use Illuminate\Http\JsonResponse;

/**
 * Every API response has the shape:
 * { success, message, data, error_code? (errors only), meta? (pagination), errors? (validation) }
 */
class ResponseHelper
{
    public static function success(StatusCode|int $statusCode, ResponseMessage|string|null $message = null, $data = null, ?array $meta = null): JsonResponse
    {
        $status = self::statusValue($statusCode);

        $body = [
            'success' => true,
            'message' => self::messageValue($message) ?? self::defaultSuccessMessage($status),
            'data'    => $data ?? (object) [],
        ];

        if ($meta !== null) {
            $body['meta'] = $meta;
        }

        return response()->json($body, $status);
    }

    public static function error(StatusCode|int $statusCode, ResponseMessage|string|null $message = null, $errors = null, ?ErrorCode $errorCode = null): JsonResponse
    {
        $status = self::statusValue($statusCode);

        $body = [
            'success'    => false,
            'message'    => self::messageValue($message) ?? ResponseMessage::INTERNAL_SERVER_ERROR->value,
            'error_code' => ($errorCode ?? ErrorCode::fromStatus($status))->value,
            'data'       => (object) [],
        ];

        if ($errors !== null) {
            $body['errors'] = $errors;
        }

        return response()->json($body, $status);
    }

    public static function responseSuccessAllData($data, $statusCode)
    {
        return self::success($statusCode, null, $data);
    }

    public static function responseSuccessDataWithPaginate($data, $page, $limit, $total, $statusCode)
    {
        return self::success($statusCode, null, $data, [
            'current_page' => (int) $page,
            'per_page'     => (int) $limit,
            'total'        => (int) $total,
            'last_page'    => (int) ceil($total / $limit),
        ]);
    }

    public static function responseSuccessSingleData($data, $statusCode)
    {
        return self::success($statusCode, null, $data);
    }

    public static function responseMessageRequest($statusCode, $isSuccess, $message)
    {
        return $isSuccess
            ? self::success($statusCode, $message)
            : self::error($statusCode, $message);
    }

    public static function responseMessageRequestWithErrors($statusCode, $isSuccess, $message, $errors)
    {
        return self::error($statusCode, $message, $errors, ErrorCode::VALIDATION_FAILED);
    }

    public static function responseMessageRequestWithData($statusCode, $isSuccess, $message, $data)
    {
        return $isSuccess
            ? self::success($statusCode, $message, $data)
            : self::error($statusCode, $message);
    }

    private static function statusValue(StatusCode|int $statusCode): int
    {
        return $statusCode instanceof StatusCode ? $statusCode->value : $statusCode;
    }

    private static function messageValue(ResponseMessage|string|null $message): ?string
    {
        return $message instanceof ResponseMessage ? $message->value : $message;
    }

    private static function defaultSuccessMessage(int $status): string
    {
        return $status === StatusCode::CREATED->value
            ? ResponseMessage::SUCCESS_CREATE->value
            : ResponseMessage::FOUND->value;
    }
}

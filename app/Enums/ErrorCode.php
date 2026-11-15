<?php
namespace App\Enums;

enum ErrorCode: int {
    /* 10XX : General */
    case SERVER_ERROR      = 1000;
    case BAD_REQUEST       = 1001;
    case TOO_MANY_REQUESTS = 1006;
    /* 11XX : HTTP */
    case UNAUTHENTICATED    = 1101;
    case FORBIDDEN          = 1102;
    case VALIDATION_FAILED  = 1103;
    case NOT_FOUND          = 1105;
    case METHOD_NOT_ALLOWED = 1106;
    /* 13XX : Session */
    case INVALID_CREDENTIALS = 1301;

    public static function fromStatus(int $status): self
    {
        return match ($status) {
            400     => self::BAD_REQUEST,
            401     => self::UNAUTHENTICATED,
            403     => self::FORBIDDEN,
            404     => self::NOT_FOUND,
            405     => self::METHOD_NOT_ALLOWED,
            422     => self::VALIDATION_FAILED,
            429     => self::TOO_MANY_REQUESTS,
            default => self::SERVER_ERROR,
        };
    }
}

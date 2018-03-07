type token =
  | EOF
  | TDot
  | TComma
  | TSemicolon
  | TPtr
  | TEqual
  | TLParen
  | TRParen
  | TLBrace
  | TRBrace
  | TLBrack
  | TRBrack
  | Tpublic
  | Tinclude
  | Tconst
  | Tidentifier of (string)
  | Tnumber of (int)
  | Tstring of (string)
  | Tchar
  | Tshort
  | Tunsigned
  | Tint
  | Tfloat
  | Tdouble
  | Tint8
  | Tint16
  | Tint32
  | Tint64
  | Tuint8
  | Tuint16
  | Tuint32
  | Tuint64
  | Tsizet
  | Twchar
  | Tvoid
  | Tlong
  | Tstruct
  | Tunion
  | Tenum
  | Tenclave
  | Tfrom
  | Timport
  | Ttrusted
  | Tuntrusted
  | Tallow
  | Tpropagate_errno

val start_parsing :
  (Lexing.lexbuf  -> token) -> Lexing.lexbuf -> Ast.enclave

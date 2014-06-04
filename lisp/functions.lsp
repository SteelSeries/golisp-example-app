(define (fact x)
  (cond ((< x 2) 1)
        (else (* x (fact (- x 1))))))

(define (double-fact x)
  (* 2 (go-fact x)))

(define (multiply x y)
  (* x y))

(define (scale x)
  (* x CONSTANT))

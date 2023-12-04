/**
 * There are several occasions where a match to check if
 * the Result is Ok or Err, this solves the verbose by
 * replacing it with one single macro.
 */
#[macro_export]
macro_rules! unwrap_or_return {
    ( $e:expr ) => {
        match $e {
            Ok(x) => x,
            Err(err) => return Err(err),
        }
    };
}

#[macro_export]
macro_rules! unwrap_or_return_with {
    ( $e:expr, $reason:expr) => {
        match $e {
            Ok(x) => x,
            Err(_) => return $reason,
        }
    };
}

pub use unwrap_or_return;
pub use unwrap_or_return_with;

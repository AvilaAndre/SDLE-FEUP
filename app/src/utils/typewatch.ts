export var typewatch = (function () {
    var timer = 0;
    return function (callback: TimerHandler, ms: number | undefined) {
        clearTimeout(timer);
        timer = setTimeout(callback, ms);
    };
})();

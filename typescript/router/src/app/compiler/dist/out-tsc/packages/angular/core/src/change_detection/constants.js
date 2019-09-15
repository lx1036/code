"use strict";
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
/**
 * Describes within the change detector which strategy will be used the next time change
 * detection is triggered.
 *
 */
/**
 * Describes within the change detector which strategy will be used the next time change
 * detection is triggered.
 *
 */
var ChangeDetectionStrategy;
/**
 * Describes within the change detector which strategy will be used the next time change
 * detection is triggered.
 *
 */
(function (ChangeDetectionStrategy) {
    /**
     * `OnPush` means that the change detector's mode will be initially set to `CheckOnce`.
     */
    ChangeDetectionStrategy[ChangeDetectionStrategy["OnPush"] = 0] = "OnPush";
    /**
     * `Default` means that the change detector's mode will be initially set to `CheckAlways`.
     */
    ChangeDetectionStrategy[ChangeDetectionStrategy["Default"] = 1] = "Default";
})(ChangeDetectionStrategy = exports.ChangeDetectionStrategy || (exports.ChangeDetectionStrategy = {}));
/**
 * Describes the status of the detector.
 */
/**
 * Describes the status of the detector.
 */
var ChangeDetectorStatus;
/**
 * Describes the status of the detector.
 */
(function (ChangeDetectorStatus) {
    /**
     * `CheckOnce` means that after calling detectChanges the mode of the change detector
     * will become `Checked`.
     */
    ChangeDetectorStatus[ChangeDetectorStatus["CheckOnce"] = 0] = "CheckOnce";
    /**
     * `Checked` means that the change detector should be skipped until its mode changes to
     * `CheckOnce`.
     */
    ChangeDetectorStatus[ChangeDetectorStatus["Checked"] = 1] = "Checked";
    /**
     * `CheckAlways` means that after calling detectChanges the mode of the change detector
     * will remain `CheckAlways`.
     */
    ChangeDetectorStatus[ChangeDetectorStatus["CheckAlways"] = 2] = "CheckAlways";
    /**
     * `Detached` means that the change detector sub tree is not a part of the main tree and
     * should be skipped.
     */
    ChangeDetectorStatus[ChangeDetectorStatus["Detached"] = 3] = "Detached";
    /**
     * `Errored` means that the change detector encountered an error checking a binding
     * or calling a directive lifecycle method and is now in an inconsistent state. Change
     * detectors in this state will no longer detect changes.
     */
    ChangeDetectorStatus[ChangeDetectorStatus["Errored"] = 4] = "Errored";
    /**
     * `Destroyed` means that the change detector is destroyed.
     */
    ChangeDetectorStatus[ChangeDetectorStatus["Destroyed"] = 5] = "Destroyed";
})(ChangeDetectorStatus = exports.ChangeDetectorStatus || (exports.ChangeDetectorStatus = {}));
function isDefaultChangeDetectionStrategy(changeDetectionStrategy) {
    return changeDetectionStrategy == null ||
        changeDetectionStrategy === ChangeDetectionStrategy.Default;
}
exports.isDefaultChangeDetectionStrategy = isDefaultChangeDetectionStrategy;
//# sourceMappingURL=constants.js.map
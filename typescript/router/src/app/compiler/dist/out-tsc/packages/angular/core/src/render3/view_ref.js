"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
Object.defineProperty(exports, "__esModule", { value: true });
var instructions_1 = require("./instructions");
var util_1 = require("./util");
var ViewRef = /** @class */ (function () {
    function ViewRef(_view, context) {
        this._view = _view;
        this.context = (context);
    }
    /** @internal */
    /** @internal */
    ViewRef.prototype._setComponentContext = /** @internal */
    function (view, context) {
        this._view = view;
        this.context = context;
    };
    ViewRef.prototype.destroy = function () { util_1.notImplemented(); };
    ViewRef.prototype.onDestroy = function (callback) { util_1.notImplemented(); };
    /**
     * Marks a view and all of its ancestors dirty.
     *
     * It also triggers change detection by calling `scheduleTick` internally, which coalesces
     * multiple `markForCheck` calls to into one change detection run.
     *
     * This can be used to ensure an {@link ChangeDetectionStrategy#OnPush OnPush} component is
     * checked when it needs to be re-rendered but the two normal triggers haven't marked it
     * dirty (i.e. inputs haven't changed and events haven't fired in the view).
     *
     * <!-- TODO: Add a link to a chapter on OnPush components -->
     *
     * ### Example
     *
     * ```typescript
     * @Component({
     *   selector: 'my-app',
     *   template: `Number of ticks: {{numberOfTicks}}`
     *   changeDetection: ChangeDetectionStrategy.OnPush,
     * })
     * class AppComponent {
     *   numberOfTicks = 0;
     *
     *   constructor(private ref: ChangeDetectorRef) {
     *     setInterval(() => {
     *       this.numberOfTicks++;
     *       // the following is required, otherwise the view will not be updated
     *       this.ref.markForCheck();
     *     }, 1000);
     *   }
     * }
     * ```
     */
    /**
       * Marks a view and all of its ancestors dirty.
       *
       * It also triggers change detection by calling `scheduleTick` internally, which coalesces
       * multiple `markForCheck` calls to into one change detection run.
       *
       * This can be used to ensure an {@link ChangeDetectionStrategy#OnPush OnPush} component is
       * checked when it needs to be re-rendered but the two normal triggers haven't marked it
       * dirty (i.e. inputs haven't changed and events haven't fired in the view).
       *
       * <!-- TODO: Add a link to a chapter on OnPush components -->
       *
       * ### Example
       *
       * ```typescript
       * @Component({
       *   selector: 'my-app',
       *   template: `Number of ticks: {{numberOfTicks}}`
       *   changeDetection: ChangeDetectionStrategy.OnPush,
       * })
       * class AppComponent {
       *   numberOfTicks = 0;
       *
       *   constructor(private ref: ChangeDetectorRef) {
       *     setInterval(() => {
       *       this.numberOfTicks++;
       *       // the following is required, otherwise the view will not be updated
       *       this.ref.markForCheck();
       *     }, 1000);
       *   }
       * }
       * ```
       */
    ViewRef.prototype.markForCheck = /**
       * Marks a view and all of its ancestors dirty.
       *
       * It also triggers change detection by calling `scheduleTick` internally, which coalesces
       * multiple `markForCheck` calls to into one change detection run.
       *
       * This can be used to ensure an {@link ChangeDetectionStrategy#OnPush OnPush} component is
       * checked when it needs to be re-rendered but the two normal triggers haven't marked it
       * dirty (i.e. inputs haven't changed and events haven't fired in the view).
       *
       * <!-- TODO: Add a link to a chapter on OnPush components -->
       *
       * ### Example
       *
       * ```typescript
       * @Component({
       *   selector: 'my-app',
       *   template: `Number of ticks: {{numberOfTicks}}`
       *   changeDetection: ChangeDetectionStrategy.OnPush,
       * })
       * class AppComponent {
       *   numberOfTicks = 0;
       *
       *   constructor(private ref: ChangeDetectorRef) {
       *     setInterval(() => {
       *       this.numberOfTicks++;
       *       // the following is required, otherwise the view will not be updated
       *       this.ref.markForCheck();
       *     }, 1000);
       *   }
       * }
       * ```
       */
    function () { instructions_1.markViewDirty(this._view); };
    /**
     * Detaches the view from the change detection tree.
     *
     * Detached views will not be checked during change detection runs until they are
     * re-attached, even if they are dirty. `detach` can be used in combination with
     * {@link ChangeDetectorRef#detectChanges detectChanges} to implement local change
     * detection checks.
     *
     * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
     * <!-- TODO: Add a live demo once ref.detectChanges is merged into master -->
     *
     * ### Example
     *
     * The following example defines a component with a large list of readonly data.
     * Imagine the data changes constantly, many times per second. For performance reasons,
     * we want to check and update the list every five seconds. We can do that by detaching
     * the component's change detector and doing a local check every five seconds.
     *
     * ```typescript
     * class DataProvider {
     *   // in a real application the returned data will be different every time
     *   get data() {
     *     return [1,2,3,4,5];
     *   }
     * }
     *
     * @Component({
     *   selector: 'giant-list',
     *   template: `
     *     <li *ngFor="let d of dataProvider.data">Data {{d}}</li>
     *   `,
     * })
     * class GiantList {
     *   constructor(private ref: ChangeDetectorRef, private dataProvider: DataProvider) {
     *     ref.detach();
     *     setInterval(() => {
     *       this.ref.detectChanges();
     *     }, 5000);
     *   }
     * }
     *
     * @Component({
     *   selector: 'app',
     *   providers: [DataProvider],
     *   template: `
     *     <giant-list><giant-list>
     *   `,
     * })
     * class App {
     * }
     * ```
     */
    /**
       * Detaches the view from the change detection tree.
       *
       * Detached views will not be checked during change detection runs until they are
       * re-attached, even if they are dirty. `detach` can be used in combination with
       * {@link ChangeDetectorRef#detectChanges detectChanges} to implement local change
       * detection checks.
       *
       * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
       * <!-- TODO: Add a live demo once ref.detectChanges is merged into master -->
       *
       * ### Example
       *
       * The following example defines a component with a large list of readonly data.
       * Imagine the data changes constantly, many times per second. For performance reasons,
       * we want to check and update the list every five seconds. We can do that by detaching
       * the component's change detector and doing a local check every five seconds.
       *
       * ```typescript
       * class DataProvider {
       *   // in a real application the returned data will be different every time
       *   get data() {
       *     return [1,2,3,4,5];
       *   }
       * }
       *
       * @Component({
       *   selector: 'giant-list',
       *   template: `
       *     <li *ngFor="let d of dataProvider.data">Data {{d}}</li>
       *   `,
       * })
       * class GiantList {
       *   constructor(private ref: ChangeDetectorRef, private dataProvider: DataProvider) {
       *     ref.detach();
       *     setInterval(() => {
       *       this.ref.detectChanges();
       *     }, 5000);
       *   }
       * }
       *
       * @Component({
       *   selector: 'app',
       *   providers: [DataProvider],
       *   template: `
       *     <giant-list><giant-list>
       *   `,
       * })
       * class App {
       * }
       * ```
       */
    ViewRef.prototype.detach = /**
       * Detaches the view from the change detection tree.
       *
       * Detached views will not be checked during change detection runs until they are
       * re-attached, even if they are dirty. `detach` can be used in combination with
       * {@link ChangeDetectorRef#detectChanges detectChanges} to implement local change
       * detection checks.
       *
       * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
       * <!-- TODO: Add a live demo once ref.detectChanges is merged into master -->
       *
       * ### Example
       *
       * The following example defines a component with a large list of readonly data.
       * Imagine the data changes constantly, many times per second. For performance reasons,
       * we want to check and update the list every five seconds. We can do that by detaching
       * the component's change detector and doing a local check every five seconds.
       *
       * ```typescript
       * class DataProvider {
       *   // in a real application the returned data will be different every time
       *   get data() {
       *     return [1,2,3,4,5];
       *   }
       * }
       *
       * @Component({
       *   selector: 'giant-list',
       *   template: `
       *     <li *ngFor="let d of dataProvider.data">Data {{d}}</li>
       *   `,
       * })
       * class GiantList {
       *   constructor(private ref: ChangeDetectorRef, private dataProvider: DataProvider) {
       *     ref.detach();
       *     setInterval(() => {
       *       this.ref.detectChanges();
       *     }, 5000);
       *   }
       * }
       *
       * @Component({
       *   selector: 'app',
       *   providers: [DataProvider],
       *   template: `
       *     <giant-list><giant-list>
       *   `,
       * })
       * class App {
       * }
       * ```
       */
    function () { this._view.flags &= ~8 /* Attached */; };
    /**
     * Re-attaches a view to the change detection tree.
     *
     * This can be used to re-attach views that were previously detached from the tree
     * using {@link ChangeDetectorRef#detach detach}. Views are attached to the tree by default.
     *
     * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
     *
     * ### Example
     *
     * The following example creates a component displaying `live` data. The component will detach
     * its change detector from the main change detector tree when the component's live property
     * is set to false.
     *
     * ```typescript
     * class DataProvider {
     *   data = 1;
     *
     *   constructor() {
     *     setInterval(() => {
     *       this.data = this.data * 2;
     *     }, 500);
     *   }
     * }
     *
     * @Component({
     *   selector: 'live-data',
     *   inputs: ['live'],
     *   template: 'Data: {{dataProvider.data}}'
     * })
     * class LiveData {
     *   constructor(private ref: ChangeDetectorRef, private dataProvider: DataProvider) {}
     *
     *   set live(value) {
     *     if (value) {
     *       this.ref.reattach();
     *     } else {
     *       this.ref.detach();
     *     }
     *   }
     * }
     *
     * @Component({
     *   selector: 'my-app',
     *   providers: [DataProvider],
     *   template: `
     *     Live Update: <input type="checkbox" [(ngModel)]="live">
     *     <live-data [live]="live"><live-data>
     *   `,
     * })
     * class AppComponent {
     *   live = true;
     * }
     * ```
     */
    /**
       * Re-attaches a view to the change detection tree.
       *
       * This can be used to re-attach views that were previously detached from the tree
       * using {@link ChangeDetectorRef#detach detach}. Views are attached to the tree by default.
       *
       * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
       *
       * ### Example
       *
       * The following example creates a component displaying `live` data. The component will detach
       * its change detector from the main change detector tree when the component's live property
       * is set to false.
       *
       * ```typescript
       * class DataProvider {
       *   data = 1;
       *
       *   constructor() {
       *     setInterval(() => {
       *       this.data = this.data * 2;
       *     }, 500);
       *   }
       * }
       *
       * @Component({
       *   selector: 'live-data',
       *   inputs: ['live'],
       *   template: 'Data: {{dataProvider.data}}'
       * })
       * class LiveData {
       *   constructor(private ref: ChangeDetectorRef, private dataProvider: DataProvider) {}
       *
       *   set live(value) {
       *     if (value) {
       *       this.ref.reattach();
       *     } else {
       *       this.ref.detach();
       *     }
       *   }
       * }
       *
       * @Component({
       *   selector: 'my-app',
       *   providers: [DataProvider],
       *   template: `
       *     Live Update: <input type="checkbox" [(ngModel)]="live">
       *     <live-data [live]="live"><live-data>
       *   `,
       * })
       * class AppComponent {
       *   live = true;
       * }
       * ```
       */
    ViewRef.prototype.reattach = /**
       * Re-attaches a view to the change detection tree.
       *
       * This can be used to re-attach views that were previously detached from the tree
       * using {@link ChangeDetectorRef#detach detach}. Views are attached to the tree by default.
       *
       * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
       *
       * ### Example
       *
       * The following example creates a component displaying `live` data. The component will detach
       * its change detector from the main change detector tree when the component's live property
       * is set to false.
       *
       * ```typescript
       * class DataProvider {
       *   data = 1;
       *
       *   constructor() {
       *     setInterval(() => {
       *       this.data = this.data * 2;
       *     }, 500);
       *   }
       * }
       *
       * @Component({
       *   selector: 'live-data',
       *   inputs: ['live'],
       *   template: 'Data: {{dataProvider.data}}'
       * })
       * class LiveData {
       *   constructor(private ref: ChangeDetectorRef, private dataProvider: DataProvider) {}
       *
       *   set live(value) {
       *     if (value) {
       *       this.ref.reattach();
       *     } else {
       *       this.ref.detach();
       *     }
       *   }
       * }
       *
       * @Component({
       *   selector: 'my-app',
       *   providers: [DataProvider],
       *   template: `
       *     Live Update: <input type="checkbox" [(ngModel)]="live">
       *     <live-data [live]="live"><live-data>
       *   `,
       * })
       * class AppComponent {
       *   live = true;
       * }
       * ```
       */
    function () { this._view.flags |= 8 /* Attached */; };
    /**
     * Checks the view and its children.
     *
     * This can also be used in combination with {@link ChangeDetectorRef#detach detach} to implement
     * local change detection checks.
     *
     * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
     * <!-- TODO: Add a live demo once ref.detectChanges is merged into master -->
     *
     * ### Example
     *
     * The following example defines a component with a large list of readonly data.
     * Imagine, the data changes constantly, many times per second. For performance reasons,
     * we want to check and update the list every five seconds.
     *
     * We can do that by detaching the component's change detector and doing a local change detection
     * check every five seconds.
     *
     * See {@link ChangeDetectorRef#detach detach} for more information.
     */
    /**
       * Checks the view and its children.
       *
       * This can also be used in combination with {@link ChangeDetectorRef#detach detach} to implement
       * local change detection checks.
       *
       * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
       * <!-- TODO: Add a live demo once ref.detectChanges is merged into master -->
       *
       * ### Example
       *
       * The following example defines a component with a large list of readonly data.
       * Imagine, the data changes constantly, many times per second. For performance reasons,
       * we want to check and update the list every five seconds.
       *
       * We can do that by detaching the component's change detector and doing a local change detection
       * check every five seconds.
       *
       * See {@link ChangeDetectorRef#detach detach} for more information.
       */
    ViewRef.prototype.detectChanges = /**
       * Checks the view and its children.
       *
       * This can also be used in combination with {@link ChangeDetectorRef#detach detach} to implement
       * local change detection checks.
       *
       * <!-- TODO: Add a link to a chapter on detach/reattach/local digest -->
       * <!-- TODO: Add a live demo once ref.detectChanges is merged into master -->
       *
       * ### Example
       *
       * The following example defines a component with a large list of readonly data.
       * Imagine, the data changes constantly, many times per second. For performance reasons,
       * we want to check and update the list every five seconds.
       *
       * We can do that by detaching the component's change detector and doing a local change detection
       * check every five seconds.
       *
       * See {@link ChangeDetectorRef#detach detach} for more information.
       */
    function () { instructions_1.detectChanges(this.context); };
    /**
     * Checks the change detector and its children, and throws if any changes are detected.
     *
     * This is used in development mode to verify that running change detection doesn't
     * introduce other changes.
     */
    /**
       * Checks the change detector and its children, and throws if any changes are detected.
       *
       * This is used in development mode to verify that running change detection doesn't
       * introduce other changes.
       */
    ViewRef.prototype.checkNoChanges = /**
       * Checks the change detector and its children, and throws if any changes are detected.
       *
       * This is used in development mode to verify that running change detection doesn't
       * introduce other changes.
       */
    function () { instructions_1.checkNoChanges(this.context); };
    return ViewRef;
}());
exports.ViewRef = ViewRef;
var EmbeddedViewRef = /** @class */ (function (_super) {
    __extends(EmbeddedViewRef, _super);
    function EmbeddedViewRef(viewNode, template, context) {
        var _this = _super.call(this, viewNode.data, context) || this;
        _this._lViewNode = viewNode;
        return _this;
    }
    return EmbeddedViewRef;
}(ViewRef));
exports.EmbeddedViewRef = EmbeddedViewRef;
/**
 * Creates a ViewRef bundled with destroy functionality.
 *
 * @param context The context for this view
 * @returns The ViewRef
 */
function createViewRef(view, context) {
    // TODO: add detectChanges back in when implementing ChangeDetectorRef.detectChanges
    return addDestroyable(new ViewRef((view), context));
}
exports.createViewRef = createViewRef;
/**
 * Decorates an object with destroy logic (implementing the DestroyRef interface)
 * and returns the enhanced object.
 *
 * @param obj The object to decorate
 * @returns The object with destroy logic
 */
function addDestroyable(obj) {
    var destroyFn = null;
    obj.destroyed = false;
    obj.destroy = function () {
        destroyFn && destroyFn.forEach(function (fn) { return fn(); });
        this.destroyed = true;
    };
    obj.onDestroy = function (fn) { return (destroyFn || (destroyFn = [])).push(fn); };
    return obj;
}
exports.addDestroyable = addDestroyable;
//# sourceMappingURL=view_ref.js.map
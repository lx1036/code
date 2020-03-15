import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {FlexLayoutModule} from '@angular/flex-layout';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {AceEditorModule} from 'ng2-ace-editor';
import {FilterPipeModule} from 'ngx-filter-pipe';
import {NgxChartsModule} from '@swimlane/ngx-charts';

import {
  MAT_TOOLTIP_DEFAULT_OPTIONS,
  MatButtonModule,
  MatButtonToggleModule,
  MatCardModule,
  MatCheckboxModule,
  MatChipsModule,
  MatDialogModule,
  MatDividerModule,
  MatFormFieldModule,
  MatGridListModule,
  MatIconModule,
  MatInputModule,
  MatMenuModule,
  MatPaginatorModule,
  MatProgressBarModule,
  MatProgressSpinnerModule,
  MatRadioModule,
  MatSelectModule,
  MatSidenavModule,
  MatSliderModule,
  MatSlideToggleModule,
  MatSnackBarModule,
  MatSortModule,
  MatTableModule,
  MatTabsModule,
  MatToolbarModule,
  MatTooltipModule,
} from '@angular/material';
import {RouterModule} from "@angular/router";
import {KD_TOOLTIP_DEFAULT_OPTIONS} from "./index.config";


const SHARED_DEPENDENCIES = [
  // Angular imports
  CommonModule,
  FormsModule,
  ReactiveFormsModule,

  // Material imports
  MatButtonModule,
  MatCardModule,
  MatCheckboxModule,
  MatDividerModule,
  MatGridListModule,
  MatFormFieldModule,
  MatIconModule,
  MatInputModule,
  MatProgressSpinnerModule,
  MatProgressBarModule,
  MatRadioModule,
  MatSidenavModule,
  MatToolbarModule,
  MatTooltipModule,
  MatSliderModule,
  MatDialogModule,
  MatSlideToggleModule,
  MatChipsModule,
  MatTableModule,
  MatSortModule,
  MatPaginatorModule,
  MatTabsModule,
  MatMenuModule,
  MatSelectModule,
  MatButtonToggleModule,
  MatSnackBarModule,

  // Other 3rd party modules
  FlexLayoutModule,
  RouterModule,
  AceEditorModule,
  FilterPipeModule,
  NgxChartsModule,

  // Custom application modules
  // PipesModule,
];

@NgModule({
  imports: SHARED_DEPENDENCIES,
  exports: SHARED_DEPENDENCIES,
  providers: [
    {
      provide: MAT_TOOLTIP_DEFAULT_OPTIONS,
      useValue: KD_TOOLTIP_DEFAULT_OPTIONS,
    },
  ],
})
export class SharedModule {}

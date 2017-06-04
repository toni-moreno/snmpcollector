/*
 * Angular 2 Dropdown Multiselect for Bootstrap
 * Current version: 0.2.0
 *
 * Simon Lindh
 * https://github.com/softsimon/angular-2-dropdown-multiselect
 */

import { NgModule, Component, Pipe, OnInit, DoCheck, AfterContentInit, HostListener, Input, ElementRef, Output, EventEmitter, forwardRef, IterableDiffers } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Observable } from 'rxjs/Rx';
import { FormsModule, NG_VALUE_ACCESSOR, ControlValueAccessor } from '@angular/forms';

@Component({
    selector: 'table-actions',
    template: `
    <div class="row well" *ngIf="editEnabled === true">
      <div style="float:left; width: auto; border-right: 1px solid #1B809E; padding: 2px 10px 0px 10px; height:27px">
        <span style="padding-top: 3px">Selected Items</span>
        <label [ngClass]="itemsSelected === 0 ? ['label label-danger'] : ['label label-success']" style="font-size: 100%">{{itemsSelected}}</label>
      </div>
      <div style="float:left; width: auto; padding: 0px 10px 0px 10px;">
        <span>Select Option:</span>
        <select class="select-pages" style="width:auto" [(ngModel)]="selectedOption" (ngModelChange)="changeOption($event)">
          <option *ngFor="let option of tableAvailableActions; let i=first" style="padding-left:2px" [selected]="i ? true : false" [ngValue]="option.content">{{option.title}}</option>
        </select>
      </div>
      <div *ngIf="selectedOption" style="float:left; width: auto; border-left: 1px solid #1B809E; padding: 0px 10px 0px 10px;">
          <span *ngIf="selectedOption.type === 'button'">Action:</span>
          <!--simple button-->
          <button *ngIf="selectedOption.type === 'button'" style="padding-top: 0px; margin-top: 0px;" (click)="applyAction(selectedOption.action,null,null)" [disabled]="itemsSelected === 0">{{selectedOption.action}}</button>
          <!--selector-->
          <ng-container *ngIf="selectedOption.type === 'selector'">
            <span *ngIf="selectedOption.type === 'selector'">Field:</span>
            <select *ngIf="selectedOption.type === 'selector'" class="select-pages" style="width:auto" [(ngModel)]="selectorSelected" (ngModelChange)="changeSelector($event)">
              <option *ngFor="let option of selectedOption.options" [ngValue]="option">{{option.title}}</option>
            </select>

          </ng-container>
      </div>
      <div *ngIf="selectorSelected" style="float:left; width: auto; border-left: 1px solid #1B809E; padding: 0px 10px 0px 10px;">
      <span>Set new value:</span>
        <select *ngIf="selectorSelected.type === 'boolean'" class="select-pages" style="width:auto" [(ngModel)]="propertySelected">
          <option *ngFor="let option of selectorSelected.options" [ngValue]="option" (ngModelChange)="changeProperty($event)">{{option}}</option>
        </select>
        <ss-multiselect-dropdown *ngIf="selectorSelected.type === 'multiselector'" [options]="selectorSelected.options" [texts]="myTexts" [settings]="mySettingsInflux" [(ngModel)]="propertySelected"></ss-multiselect-dropdown>
        <div [formGroup]="selectorSelected.options" *ngIf="selectorSelected.type === 'input'" style="border:none; display:inline">
              <input formControlName="formControl" id="formControl" [(ngModel)]="propertySelected" style="width:auto"/>
              <control-messages style="display:block; margin-top:5px" [control]="selectorSelected.options.controls.formControl" ></control-messages>
        </div>
      </div>
      <div class="col-md-1" *ngIf="propertySelected">
          <button style="padding-top: 0px; margin-top: 0px;" (click)="applyAction(selectedOption.action,selectorSelected.title,propertySelected)" [disabled]="itemsSelected === 0 || (selectorSelected.type === 'input' ? !selectorSelected.options.valid : false)">Apply</button>
      </div>
      <div class="col-md-3 text-right" *ngIf="actionApplied">
      <progressbar *ngIf="actionApplied" class="progress-striped active" [value]="itemsApplied === counterErrors.length ? itemsApplied : counterItems" [max]="itemsApplied" [type]="counterItems === itemsApplied ? 'success' : 'danger'" style="width : auto">
        {{counterItems}} <i *ngIf="counterItems !== itemsApplied && counterErrors.length > 0" class="glyphicon glyphicon-exclamation-sign" [tooltip]="errorTooltip"></i> / {{itemsApplied}}
      </progressbar>
      <template #errorTooltip>
        <p *ngFor="let error of counterErrors">{{error.ID}} - {{error.error}}</p>
      </template>
      </div>
    </div>
  `,
  styleUrls: ['../css/component-styles.css']
})
export class TableActions {
    @Input() counterErrors : any = [];
    @Input() editEnabled : boolean = false;
    @Input() itemsSelected : number = 0;
    @Input() tableAvailableActions: Array<any>;
    @Input() counterItems;
    @Output() public actionApply:EventEmitter<any> = new EventEmitter<any>();

    itemsApplied: any;
    actionApplied : any = null;
    selectedOption : any = null;
    selectorSelected : any = null;
    propertySelected : any = null; //OK

    changeOption(id){
      this.actionApplied = null;
      this.propertySelected = null;
      this.selectorSelected = null;
    }

    changeSelector(id){
      this.actionApplied = null;
      this.propertySelected = null;
    }

    public applyAction(action,field?,value?): void {
      let r = true;
      r = confirm("Doing "+action + (field ? ': '+ field : '') + (value ? '='+ value : '') + " on "+this.itemsSelected+" items. Proceed?");
      if (r == true) {
        this.itemsApplied = this.itemsSelected;
        this.itemsSelected = null;
        this.actionApplied = action;
        this.actionApply.emit({action, field, value});
      }
    }

    constructor() {
    }

}

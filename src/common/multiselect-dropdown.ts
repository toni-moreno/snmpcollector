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

declare var _:any;

const MULTISELECT_VALUE_ACCESSOR: any = {
    provide: NG_VALUE_ACCESSOR,
    useExisting: forwardRef(() => MultiselectDropdown),
    multi: true
};

export interface IMultiSelectOption {
    id: any;
    name: string;
}

export interface IMultiSelectSettings {
    pullRight?: boolean;
    enableSearch?: boolean;
    checkedStyle?: 'checkboxes' | 'glyphicon';
    buttonClasses?: string;
    selectionLimit?: number;
    closeOnSelect?: boolean;
    showCheckAll?: boolean;
    showUncheckAll?: boolean;
    dynamicTitleMaxItems?: number;
    maxHeight?: string;
    singleSelect?: boolean;
}

export interface IMultiSelectTexts {
    checkAll?: string;
    uncheckAll?: string;
    checked?: string;
    checkedPlural?: string;
    searchPlaceholder?: string;
    defaultTitle?: string;
}

@Pipe({
    name: 'searchFilter'
})
export class MultiSelectSearchFilter {
    transform(options: Array<IMultiSelectOption>, args: string): Array<IMultiSelectOption> {
        return options.filter((option: IMultiSelectOption) => option.name.toLowerCase().indexOf((args || '').toLowerCase()) > -1);
    }
}

@Component({
    selector: 'ss-multiselect-dropdown',
    providers: [MULTISELECT_VALUE_ACCESSOR],
    styles: [`
		a { outline: none !important; }
	`],
    template: `
        <div class="btn-group">
            <button type="button" class="dropdown-toggle" [ngClass]="settings.buttonClasses" (click)="toggleDropdown()">{{ title }}&nbsp;<span class="caret"></span></button>
            <ul *ngIf="isVisible" class="dropdown-menu" [class.pull-right]="settings.pullRight" [style.max-height]="settings.maxHeight" style="display: block; height: auto; overflow-y: auto; z-index:1001; min-width:300px">
                <li style="margin: 0px 5px 5px 5px;" *ngIf="settings.enableSearch">
                    <div class="input-group input-group-sm">
                        <span class="input-group-addon" id="sizing-addon3" (click)="clearSearch()" role="button"><i class="glyphicon glyphicon-trash"></i></span>
                        <input type="text" class="form-control" placeholder="{{ texts.searchPlaceholder }}" aria-describedby="sizing-addon3" [(ngModel)]="searchFilterText">
                        <span class="input-group-btn">
                            <button class="btn btn-default" role="text"><label style="font-size: 100%" class="label label-info">{{(options|searchFilter:searchFilterText).length}} results</label></button>
                        </span>
                    </div>
                </li>
                <li class="divider" *ngIf="settings.enableSearch"></li>
                <li *ngIf="settings.showCheckAll && settings.singleSelect === false">
                    <a href="javascript:;" role="menuitem" tabindex="-1" (click)="checkAll()">
                        <span style="width: 16px;" class="glyphicon glyphicon-ok text-success"></span>
                        {{ texts.checkAll }}
                    </a>
                </li>
                <li *ngIf="settings.showUncheckAll && settings.singleSelect === false">
                    <a href="javascript:;" role="menuitem" tabindex="-1" (click)="uncheckAll()">
                        <span style="width: 16px;" class="glyphicon glyphicon-remove text-danger"></span>
                        {{ texts.uncheckAll }}
                    </a>
                </li>
                <li *ngIf="(settings.showCheckAll || settings.showUncheckAll) && settings.singleSelect === false" class="divider"></li>
                <li *ngFor="let option of options | searchFilter:searchFilterText">
                    <a href="javascript:;" role="menuitem" tabindex="-1" (click)="setSelected($event, option)">
                        <input *ngIf="settings.checkedStyle == 'checkboxes'" type="checkbox" [checked]="isSelected(option)" />
                        <span *ngIf="settings.checkedStyle == 'glyphicon'" style="width: 16px;" [ngClass]="isSelected(option) ? ['glyphicon glyphicon-ok' , 'text-success'] : 'glyphicon'"></span>
                        {{ option.name }}
                    </a>
                </li>
            </ul>
        </div>
    `
})
export class MultiselectDropdown implements OnInit, DoCheck, ControlValueAccessor {

    @Input() options: Array<IMultiSelectOption>;
    @Input() settings: IMultiSelectSettings;
    @Input() texts: IMultiSelectTexts;
    @Output() selectionLimitReached = new EventEmitter();
    @HostListener('document: click', ['$event.target'])
    onClick(target) {
        let parentFound = false;
        while (target !== null && !parentFound) {
            if (target === this.element.nativeElement) {
                parentFound = true;
            }
            target = target.parentElement;
        }
        if (!parentFound) {
            this.isVisible = false;
        }
    }

    protected onModelChange: Function = (_: any) => {};
    protected onModelTouched: Function = () => {};
    protected model: any[];
    protected title: string;
    protected differ: any;
    protected numSelected: number = 0;
    protected isVisible: boolean = false;
    protected searchFilterText: string = '';
    protected defaultSettings: IMultiSelectSettings = {
        pullRight: false,
        enableSearch: true,
        checkedStyle: 'glyphicon',
        buttonClasses: 'btn btn-default',
        selectionLimit: 0,
        closeOnSelect: false,
        showCheckAll: true,
        showUncheckAll: true,
        dynamicTitleMaxItems: 3,
        maxHeight: '300px',
        singleSelect: false,
    };
    protected defaultTexts: IMultiSelectTexts = {
        checkAll: 'Check all filtered',
        uncheckAll: 'Uncheck all',
        checked: 'checked',
        checkedPlural: 'checked',
        searchPlaceholder: 'Search...',
        defaultTitle: 'Select',
    };

    constructor(
        protected element: ElementRef,
        protected differs: IterableDiffers
    ) {
        this.differ = differs.find([]).create(null);
    }

    ngOnInit() {
        this.settings = Object.assign(this.defaultSettings, this.settings);
        this.texts = Object.assign(this.defaultTexts, this.texts);
        this.title = this.texts.defaultTitle;
    }

    writeValue(value: any) : void {
        if (value !== undefined) {
            this.model = value;
        }
    }

    registerOnChange(fn: Function): void {
        this.onModelChange = fn;
    }

    registerOnTouched(fn: Function): void {
        this.onModelTouched = fn;
    }

    ngAfterContentInit() {
        this.ngDoCheck();
    }

    ngDoCheck() {
        if(this.model){
            if (this.settings.singleSelect) {
                this.updateNumSelected();
                this.updateTitle();
            } else {
                this.updateNumSelected();
                this.updateTitle();
            }
        } else {
            this.title = this.texts.defaultTitle;
        }
    }

    clearSearch() {
        this.searchFilterText = '';
    }

    toggleDropdown() {
        this.isVisible = !this.isVisible;
    }

    isSelected(option: IMultiSelectOption): boolean {
        if (this.settings.singleSelect === true) return (this.model === option.id);
        return this.model && this.model.indexOf(option.id) > -1;
    }

    setSelected(event: Event, option: IMultiSelectOption) {

        if (this.settings.singleSelect === true) {
            this.model = option.id;
        } else {
            if (!this.model) this.model = [];
            var index = this.model.indexOf(option.id);
            if (index > -1) {
                this.model.splice(index, 1);
            } else {
                if (this.settings.selectionLimit === 0 || this.model.length < this.settings.selectionLimit) {
                    this.model.push(option.id);
                } else {
                    this.selectionLimitReached.emit(this.model.length);
                    return;
                }
            }
        }
        if (this.settings.closeOnSelect) {
        this.toggleDropdown();
        }

        this.onModelChange(this.model);
    }

    updateNumSelected() {
        if (this.settings.singleSelect) this.numSelected = 1;
        else this.numSelected = this.model && this.model.length || 0;
    }

    updateTitle() {
        if (this.numSelected === 0) {
            this.title = this.texts.defaultTitle;
        } else if (this.settings.dynamicTitleMaxItems >= this.numSelected) {
            if (this.settings.singleSelect === true) {
                this.title = this.model.toString();
            } else {
                this.title = this.options
                .filter((option: IMultiSelectOption) => this.model && this.model.indexOf(option.id) > -1)
                .map((option: IMultiSelectOption) => option.name)
                .join(', ');
            }
        } else {
            this.title = this.numSelected + ' ' + (this.numSelected === 1 ? this.texts.checked : this.texts.checkedPlural);
        }
    }

    checkAll() {
        if(!this.model) this.model = [];
        let newEntries = _.differenceWith(new MultiSelectSearchFilter().transform(this.options,this.searchFilterText).map(option => option.id),this.model,_.isEqual);
        this.model = newEntries.concat(this.model);
        this.onModelChange(this.model);
    }

    checkAllFiltered() {
        this.onModelChange(this.model);
    }

    uncheckAll() {
        this.model = [];
        this.onModelChange(this.model);
    }
}

@NgModule({
    imports: [CommonModule, FormsModule],
    exports: [MultiselectDropdown],
    declarations: [MultiselectDropdown, MultiSelectSearchFilter],
})
export class MultiselectDropdownModule { }

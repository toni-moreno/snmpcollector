import { Injectable, ApplicationRef, ViewContainerRef, Component, ComponentRef, ComponentFactoryResolver, ComponentFactory, ViewChild } from '@angular/core';

import { BlockUIComponent } from './blockui-component'; // error here, wrong path


@Injectable()
export class BlockUIService {
    blockComp: ComponentRef<any>;

    constructor(private _appRef: ApplicationRef, private _resolver: ComponentFactoryResolver) {
    }

    public start(placeholder, prop) { // placeholder missing!
        //let elementRef: ViewContainerRef = (<any>this._appRef)['_rootComponents'][0].location; // remove this
        let elementRef = placeholder; // add this
        return this.startInside(elementRef, null, prop);
    }

    public setProperty(prop) {
      this.blockComp.instance.state.message = prop;
    }

    public startInside(elementRef: ViewContainerRef, anchorName: string, prop? : string) {
        let factory = this._resolver.resolveComponentFactory(BlockUIComponent);
        this.blockComp = elementRef.createComponent(factory);
        prop ? this.blockComp.instance.message = prop : 'Please wait...';
    }

    public stop() {
        if (this.blockComp) {
            this.blockComp.destroy();
        }
    }
}

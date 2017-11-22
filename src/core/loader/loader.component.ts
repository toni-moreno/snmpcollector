import { Component, OnInit, OnDestroy } from '@angular/core';
import { Subscription } from 'rxjs/Subscription';

import { LoaderService } from './loader.service';
import { LoaderState } from './loader';

@Component({
    selector: 'angular-loader',
    templateUrl: './loader.component.html',
    styleUrls: ['./loader.component.css']
})
export class LoaderComponent implements OnInit {

    show = false;
    message = "";
    type = "success";

    public myAlerts : Array<any> = [];

    private subscription: Subscription;
    public alertExpanded : any = false;

    constructor(
        private loaderService: LoaderService
    ) { }

    ngOnInit() {
        this.subscription = this.loaderService.loaderState
            .subscribe((state: LoaderState) => {
              console.log("STATE",state);
                this.show = state.show;
                if (state.message) {
                    this.message = state.message;
                    this.type = state.type;
                    this.myAlerts.push({'message': state.message , 'type': state.type});
                    console.log(this.myAlerts);

                }

            });
    }
    ExpandAlerts() {
        this.alertExpanded = !this.alertExpanded;
    }

    ngOnDestroy() {
        this.subscription.unsubscribe();
    }
}

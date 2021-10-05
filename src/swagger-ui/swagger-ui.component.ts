import { Component, OnInit } from '@angular/core';
import SwaggerUI from "swagger-ui"

declare const SwaggerUIBundle: any;

@Component({
  selector: 'app-swagger-ui',
  templateUrl: './swagger-ui.component.html',
  styleUrls: ['./swagger-ui.component.css']
})
export class SwaggerUiComponent implements OnInit {

  ngOnInit(): void {
    
    const ui = SwaggerUI({
      dom_id: '#swagger-ui',
      layout: 'BaseLayout',
      presets: [
        SwaggerUI.presets.apis,
      ],
      url: '../assets/swagger.yaml',
      docExpansion: 'none',
      operationsSorter: 'alpha'
    });
   }

}

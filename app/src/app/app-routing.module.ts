import { NgModule } from '@angular/core';
import { MesComponent } from './mes.component';
import { RouterModule, Routes } from '@angular/router';
import { MultiRouterService } from './framework/multirouter.service';
import { GeneralComponent } from './general.component';
import { BeeComponent } from './bee.component';
import { Bee1Component } from './bee1.component';

const routes: Routes = [
  {
    path: '',
    component: MesComponent
  },
  {
    path: 'home',
    component: MesComponent
  },
  {
    path: 'general',
    component: GeneralComponent,
    pathMatch: 'prefix',
    children: [
      {
        path: 'bee',
        component: BeeComponent
      },
      {
        path: 'bee1/:channel',
        component: Bee1Component
      }
    ]
  }
];


@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
  providers: [MultiRouterService]
})
export class AppRoutingModule { }
